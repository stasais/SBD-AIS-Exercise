# System Architecture - Detailed Analysis

## Overview Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              BROWSER (Client)                                    │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                    Frontend (index.html)                                 │    │
│  │  fetch("http://orders.localhost/api/menu")                              │    │
│  │  fetch("http://orders.localhost/api/order", {POST})                     │    │
│  │  fetch("http://orders.localhost/api/order/all")                         │    │
│  │  fetch("http://orders.localhost/api/order/totalled")                    │    │
│  │  fetch("http://orders.localhost/api/receipt/{id}")                      │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
└───────────────────────────────────┬─────────────────────────────────────────────┘
                                    │ HTTP (port 80)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           TRAEFIK (Reverse Proxy)                               │
│                                                                                  │
│  Routes:                                                                         │
│  - Host: orders.localhost  →  orderservice:3000                                 │
│  - PathPrefix: /           →  frontend:80                                       │
│  - PathPrefix: /dashboard  →  traefik:8080 (internal)                           │
└───────────────────────────────────┬─────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
┌───────────────────────────────┐   ┌───────────────────────────────┐
│     Frontend (nginx:80)       │   │   Orderservice (Go:3000)      │
│                               │   │                               │
│  Serves static index.html     │   │  Chi Router + CORS            │
│  Network: web                 │   │  Networks: web, intercom      │
└───────────────────────────────┘   └───────────────┬───────────────┘
                                                    │
                                    ┌───────────────┴───────────────┐
                                    ▼                               ▼
                    ┌───────────────────────────┐   ┌───────────────────────────┐
                    │   PostgreSQL (:5555)      │   │     MinIO (:8500)         │
                    │                           │   │                           │
                    │  Tables:                  │   │  Bucket: "orders"         │
                    │  - drinks                 │   │  Files: order_1.md, ...   │
                    │  - orders                 │   │                           │
                    │  Network: intercom        │   │  Network: intercom        │
                    └───────────────────────────┘   └───────────────────────────┘
```

---

## 1. Frontend → API Connection

**File:** `frontend/index.html` (JavaScript)

The frontend is a **static HTML page** served by nginx. It uses **JavaScript fetch()** to call the API:

```javascript
// All API calls go to orders.localhost (routed by Traefik to orderservice)

// 1. GET menu (drinks list)
fetch("http://orders.localhost/api/menu")

// 2. POST new order
fetch("http://orders.localhost/api/order", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ drink_id: 1, amount: 2 })
})

// 3. GET all orders
fetch("http://orders.localhost/api/order/all")

// 4. GET totalled orders (aggregated)
fetch("http://orders.localhost/api/order/totalled")

// 5. GET receipt file (markdown)
fetch("http://orders.localhost/api/receipt/{orderId}")
```

**Key Point:** Frontend uses `http://orders.localhost` which:
- Resolves to your machine (localhost)
- Traefik sees `Host: orders.localhost` header
- Traefik routes to `orderservice:3000`

---

## 2. Orderservice (Go Backend)

**File:** `main.go`

```go
func main() {
    // 1. Connect to MinIO (S3)
    s3, err := storage.CreateS3client()
    
    // 2. Connect to PostgreSQL
    db, err := repository.NewDatabaseHandler()
    
    // 3. Prepopulate sample data
    repository.Prepopulate(db, s3)
    
    // 4. Setup Chi router with CORS
    r := chi.NewRouter()
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins: []string{"http://localhost", "http://localhost:3000"},
        // ...
    }))
    
    // 5. Define routes
    r.Get("/api/menu", rest.GetMenu(db))
    r.Get("/api/order/all", rest.GetOrders(db))
    r.Get("/api/order/totalled", rest.GetOrdersTotal(db))
    r.Get("/api/receipt/{orderId}", rest.GetReceiptFile(db, s3))
    r.Post("/api/order", rest.PostOrder(db, s3))
    
    // 6. Listen on port 3000
    http.ListenAndServe(":3000", r)
}
```

---

## 3. PostgreSQL Connection

**File:** `repository/db.go`

```go
func NewDatabaseHandler() (*DatabaseHandler, error) {
    dsn, err := getDsn()  // Build connection string
    dbConn, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
    
    // Auto-create tables
    dbConn.AutoMigrate(&model.Drink{}, &model.Order{})
    return &DatabaseHandler{dbConn: dbConn}, nil
}

func getDsn() (string, error) {
    // Load credentials from secrets or env vars
    dbUser, _ := secrets.LoadSecretOrEnv("POSTGRES_USER")      // → "docker"
    dbPw, _ := secrets.LoadSecretOrEnv("POSTGRES_PASSWORD")    // → "docker"
    dbName := os.LookupEnv("POSTGRES_DB")                       // → "order"
    dbPort := os.LookupEnv("PGPORT")                            // → "5555"
    dbHost := os.LookupEnv("DB_HOST")                           // → "postgres"
    
    // DSN: "host=postgres user=docker password=docker dbname=order port=5555"
    return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", ...)
}
```

**Database Tables (GORM auto-migrated):**

| Table | Fields |
|-------|--------|
| `drinks` | id, name, price, description |
| `orders` | id, created_at, amount, drink_id (FK) |

---

## 4. MinIO (S3) Connection

**File:** `storage/s3.go`

```go
func CreateS3client() (*minio.Client, error) {
    // Load from environment/secrets
    s3Endpoint := os.LookupEnv("S3_ENDPOINT")                           // → "minio:8500"
    s3AccessKeyId, _ := secrets.LoadSecretOrEnv("S3_ACCESS_KEY_ID")     // → "root"
    secretAccessKey, _ := secrets.LoadSecretOrEnv("S3_SECRET_ACCESS_KEY") // → "verysecret"
    
    // Create MinIO client
    client, err := minio.New(s3Endpoint, &minio.Options{
        Secure: false,
        Creds: credentials.NewStaticV4(s3AccessKeyId, secretAccessKey, ""),
    })
    
    // Health check (wait up to 10 seconds)
    // Create "orders" bucket if not exists
    client.MakeBucket(context.Background(), "orders", ...)
    
    return client, nil
}
```

**S3 Usage:**
- **Bucket:** `orders`
- **Files:** `order_1.md`, `order_2.md`, ... (receipt files)
- When order is created → markdown receipt saved to S3
- When receipt downloaded → read from S3 and serve

---

## 5. Secrets Loading

**File:** `secrets/load.go`

```go
func LoadSecretOrEnv(envKey string) (string, error) {
    // Try direct env var first: POSTGRES_USER=docker
    envVal, ok := os.LookupEnv(envKey)
    if ok {
        return envVal, nil
    }
    
    // Try file-based secret: POSTGRES_USER_FILE=/run/secrets/postgres_user
    envVal, ok = os.LookupEnv(envKey + "_FILE")
    if !ok {
        return "", errors.New("not set")
    }
    
    // Read secret from file
    fileContent, err := os.ReadFile(envVal)  // → reads "docker" from file
    return string(fileContent), nil
}
```

---

## 6. Data Flow Example: Place Order

```
1. User clicks "Submit Order" in browser
   ↓
2. Frontend JS: fetch("http://orders.localhost/api/order", {POST, body: {drink_id:1, amount:2}})
   ↓
3. Traefik receives request, sees Host: orders.localhost
   ↓
4. Traefik routes to orderservice:3000
   ↓
5. Orderservice rest.PostOrder():
   - Parse JSON body
   - Validate drink exists in PostgreSQL
   - Insert order into PostgreSQL
   - Generate markdown receipt
   - Upload receipt to MinIO S3
   - Return success JSON
   ↓
6. Frontend receives response, updates charts
```

---

## Summary Table

| Component | Port | Connects To | Credentials Source |
|-----------|------|-------------|-------------------|
| Frontend | 80 | Traefik (via browser) | None |
| Traefik | 80 | Frontend:80, Orderservice:3000 | None |
| Orderservice | 3000 | PostgreSQL:5555, MinIO:8500 | Docker Secrets |
| PostgreSQL | 5555 | - | `POSTGRES_USER_FILE`, `POSTGRES_PASSWORD_FILE` |
| MinIO | 8500 | - | `MINIO_ROOT_USER_FILE`, `MINIO_ROOT_PASSWORD_FILE` |

---

## Networks

| Network | Type | Services | Purpose |
|---------|------|----------|---------|
| `web` | overlay | Traefik, Frontend, Orderservice | External traffic routing |
| `intercom` | overlay | Orderservice, PostgreSQL, MinIO | Internal service communication |
