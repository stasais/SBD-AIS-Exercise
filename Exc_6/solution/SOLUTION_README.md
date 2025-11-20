# Exercise 6 Solution - S3 Object Storage Integration

## Accessible Endpoints

- **OpenAPI Documentation**: http://orders.localhost/openapi/index.html
- **Menu API**: http://orders.localhost/api/menu
- **Orders API**: http://orders.localhost/api/order
- **Receipts API**: http://orders.localhost/api/receipt/{orderId}
- **Minio Console**: http://localhost:8500 (credentials: `root` / `verysecret`)
- **Traefik Dashboard**: http://localhost:8080

## What Was Implemented

This solution adds **S3-compatible object storage** using Minio to store order receipts as markdown files. When a customer places an order, the system:

1. Saves the order to PostgreSQL database
2. Generates a markdown receipt with order details
3. Stores the receipt in Minio S3 bucket
4. Allows retrieval of receipts via REST API

## Changes Made Compared to Skeleton

### 1. `docker-compose.yml`
**What**: Added Minio service configuration  
**Why**: Provides S3-compatible object storage for receipt files  
**Where**: Lines with `order-minio` service and `order_minio_vol` volume

```yaml
volumes:
  order_minio_vol:  # Added storage volume for Minio data persistence

services:
  order-minio:
    image: minio/minio:latest
    ports:
      - "8500:8500"
    networks:
      - intercom
    command: ["server", "--address", ":8500", "/data"]
    volumes:
      - order_minio_vol:/data
    environment:
      - MINIO_ROOT_USER=root
      - MINIO_ROOT_PASSWORD=verysecret
```

Also updated Traefik to `traefik:latest` for Docker API 1.44+ compatibility.

### 2. `model/order.go`
**What**: Created markdown template for receipts  
**Why**: Formats order data into human-readable receipt format  
**Where**: `markdownTemplate` constant

```go
markdownTemplate = `# Order: %d

| Created At      | Drink ID | Amount |
|-----------------|----------|--------|
| %s | %d        | %d      |

Thanks for drinking with us!
`
```

The `ToMarkdown()` method uses `fmt.Sprintf` to populate the template with order ID, timestamp, drink ID, and amount.

### 3. `rest/api.go`
**What**: Added S3 operations for storing and retrieving receipts  
**Why**: Implements the complete receipt lifecycle - create on order placement, retrieve on demand  
**Where**: `PostOrder` and `GetReceiptFile` functions

#### In `PostOrder`:
```go
// Convert order to markdown
markdownContent := dbOrder.ToMarkdown()
reader := strings.NewReader(markdownContent)

// Upload to S3
_, err = s3.PutObject(r.Context(), "orders", dbOrder.GetFilename(), 
    reader, int64(len(markdownContent)), 
    minio.PutObjectOptions{ContentType: "text/markdown"})
```

#### In `GetReceiptFile`:
```go
// Download from S3
object, err := s3.GetObject(r.Context(), "orders", order.GetFilename(), 
    minio.GetObjectOptions{})

// Set proper HTTP headers
w.Header().Set("Content-Type", "text/markdown")
w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", order.GetFilename()))

// Stream to client
io.Copy(w, object)
```

Also added imports for `fmt` and `strings` packages.

### 4. `scripts/build-application.sh`
**What**: Added `go mod tidy` before building  
**Why**: Ensures all dependencies are properly downloaded and go.sum is updated  
**Where**: Before `go mod download`

## How It Works

1. **Order Creation Flow**:
   - Client sends POST request with drink_id and amount
   - Order is saved to PostgreSQL
   - Markdown receipt is generated using the template
   - Receipt is uploaded to Minio S3 bucket "orders"
   - Success response returned to client

2. **Receipt Retrieval Flow**:
   - Client sends GET request with order ID
   - System verifies order exists in database
   - Receipt file is fetched from S3 using order filename pattern
   - File is streamed back with proper markdown content-type headers

3. **Storage Architecture**:
   - **PostgreSQL**: Structured order data with relationships
   - **Minio S3**: Unstructured receipt documents for archival/download
   - **Docker volumes**: Persistent storage for both databases

## Testing

```bash
# Create an order
curl -X POST -H "Host: orders.localhost" http://localhost/api/order \
  -H "Content-Type: application/json" \
  -d '{"drink_id": 1, "amount": 3}'

# Retrieve receipt (replace 1 with actual order ID)
curl -H "Host: orders.localhost" http://localhost/api/receipt/1
```

Expected output:
```markdown
# Order: 1

| Created At      | Drink ID | Amount |
|-----------------|----------|--------|
| Nov 20 23:26:25 | 1        | 3      |

Thanks for drinking with us!
```
