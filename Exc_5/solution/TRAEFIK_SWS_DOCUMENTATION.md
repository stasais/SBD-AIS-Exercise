# Part 1: Traefik Reverse Proxy Configuration

## What Was Done
Configured Traefik as a reverse proxy/load balancer in the Docker Compose setup to route HTTP traffic to different microservices based on domain names.

## How It Was Implemented

### 1. Docker Compose Service Configuration
Added a complete Traefik service configuration in `docker-compose.yml` with the following components:

#### Image
```yaml
image: "traefik:v3.5.2"
```
- Uses the official Traefik v3.5.2 Docker image
- This is a stable, production-ready version of Traefik

#### Command-Line Arguments
```yaml
command:
  - "--api.insecure=true"
  - "--providers.docker=true"
  - "--providers.docker.exposedbydefault=false"
  - "--entrypoints.web.address=:80"
```

**Breakdown of each command:**

1. `--api.insecure=true`
   - **What:** Enables the Traefik web dashboard/API
   - **Why:** Provides a visual interface to see routes, services, and health status
   - **Note:** "insecure" means no authentication - suitable for development only
   - **Access:** Dashboard available at http://localhost:8080

2. `--providers.docker=true`
   - **What:** Enables Docker as a configuration provider
   - **Why:** Allows Traefik to automatically discover Docker containers
   - **How it works:** Traefik watches Docker events and reads container labels to configure routing

3. `--providers.docker.exposedbydefault=false`
   - **What:** Disables automatic exposure of all Docker containers
   - **Why:** Security best practice - services must explicitly opt-in via labels
   - **Result:** Only containers with `traefik.enable=true` label will be routed

4. `--entrypoints.web.address=:80`
   - **What:** Defines an entry point named "web" listening on port 80
   - **Why:** Sets up the main HTTP entry point for incoming traffic
   - **Entry points:** Think of these as "doors" through which traffic enters Traefik

#### Port Mappings
```yaml
ports:
  - "80:80"
  - "8080:8080"
```

**Port 80:80**
- **What:** Maps host port 80 to container port 80
- **Why:** This is the main HTTP entry point for all application traffic
- **Usage:** All HTTP requests to localhost or *.localhost go through this port

**Port 8080:8080**
- **What:** Maps host port 8080 to container port 8080
- **Why:** Exposes the Traefik dashboard for monitoring and debugging
- **Access:** Visit http://localhost:8080 to see the dashboard

#### Volume Mount
```yaml
volumes:
  - "/var/run/docker.sock:/var/run/docker.sock:ro"
```

**What this does:**
- Mounts the Docker socket into the Traefik container
- `:ro` means read-only access

**Why it's needed:**
- Traefik needs to communicate with the Docker daemon
- Allows Traefik to:
  - Discover running containers
  - Read container labels
  - Monitor container lifecycle (start/stop events)
  - Automatically update routing configuration

**Security consideration:**
- Docker socket access is powerful - only trusted containers should have it
- Read-only mode limits potential security risks

#### Network Configuration
```yaml
networks:
  - web
```

**What:**
- Connects Traefik only to the `web` network

**Why:**
- Separates concerns: Traefik handles external routing, not internal services
- Security: Traefik doesn't need access to the `intercom` network where the database resides
- Services that need to be routed by Traefik must also be on the `web` network

## Why This Approach?

### 1. **Dynamic Service Discovery**
- No need to manually configure routing for each service
- Add/remove services without restarting Traefik
- Configuration lives with the service (via labels)

### 2. **Simplified Development**
- Use domain-based routing (orders.localhost) instead of ports
- No port conflicts between services
- Easy to remember URLs

### 3. **Production-Ready Pattern**
- Same approach works in production with real domains
- Easy to add SSL/TLS certificates later
- Scalable architecture (load balancing built-in)

### 4. **Security by Default**
- Services must opt-in to be exposed
- Network isolation between frontend and backend layers
- Centralized access control point

## How Traefik Works

```
[Browser] 
    ↓
    HTTP Request to http://orders.localhost
    ↓
[Traefik on port 80] ← Reads Docker container labels
    ↓
    Matches Host(`orders.localhost`) rule
    ↓
    Routes to orderservice:3000
    ↓
[OrderService Container]
```

---

# Part 2: OrderService Traefik Labels Configuration

## What Was Done
Added Traefik labels to the `orderservice` container to enable automatic routing through Traefik, making the order service API accessible at `http://orders.localhost`.

## How It Was Implemented

### Docker Compose Labels Configuration
Added the following labels to the `orderservice` section in `docker-compose.yml`:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.orderservice.rule=Host(`orders.localhost`)"
  - "traefik.http.routers.orderservice.entrypoints=web"
  - "traefik.http.services.orderservice.loadbalancer.server.port=3000"
```

### Breakdown of Each Label

#### 1. `traefik.enable=true`
**What:**
- Explicitly enables Traefik routing for this specific container

**Why:**
- Required because we set `--providers.docker.exposedbydefault=false` in Traefik configuration
- This is an opt-in mechanism for security - only services that explicitly enable Traefik will be routed

**Without this:**
- Traefik would ignore this container completely
- No routing rules would be created

#### 2. `traefik.http.routers.orderservice.rule=Host(\`orders.localhost\`)`
**What:**
- Creates a router named "orderservice"
- Defines a routing rule based on the HTTP Host header
- Matches requests to `orders.localhost`

**Why:**
- Allows domain-based routing instead of port-based
- More intuitive URLs for development and production
- Enables multiple services to coexist on the same port (80)

**How it works:**
- When a request arrives at Traefik with Host header `orders.localhost`
- Traefik matches this rule and forwards the request to the orderservice

**Syntax explanation:**
- `traefik.http.routers` - Defines HTTP router configuration
- `.orderservice` - The unique name for this router
- `.rule` - The matching rule to apply
- `Host(\`...\`)` - Match based on the Host header value

#### 3. `traefik.http.routers.orderservice.entrypoints=web`
**What:**
- Specifies which Traefik entry point this router should use
- Links to the "web" entry point we defined in Part 1

**Why:**
- Entry points define where Traefik listens for traffic (port 80 in our case)
- A router must be associated with at least one entry point
- Allows different services to use different entry points (e.g., HTTP vs HTTPS)

**Connection to Part 1:**
- In Part 1, we defined: `--entrypoints.web.address=:80`
- This label tells the router to use that entry point

#### 4. `traefik.http.services.orderservice.loadbalancer.server.port=3000`
**What:**
- Tells Traefik which port the service is listening on inside the container
- Creates a service (backend) definition for load balancing

**Why:**
- The orderservice application listens on port 3000 internally
- Traefik needs to know where to forward the traffic
- Without this, Traefik wouldn't know which container port to use

**Important distinction:**
- This is the **container** port (3000), not a host port
- No port mapping in the `ports:` section is needed
- Traefik communicates with the service over the Docker network

**Load balancer context:**
- Even with a single instance, Traefik uses its load balancing mechanism
- If you scale to multiple instances, Traefik automatically load balances between them
- Syntax: `traefik.http.services.<service-name>.loadbalancer.server.port`

## Network Configuration

The orderservice is connected to two networks:
```yaml
networks:
  - intercom  # For database communication
  - web       # For Traefik routing
```

**Why both networks:**
- `web` network: Required for Traefik to discover and route to this service
- `intercom` network: Required for the orderservice to communicate with PostgreSQL
- Network isolation: The database is NOT on the web network (security)

## How the Routing Works

### Request Flow:
```
[Browser]
    ↓ 
    HTTP GET http://orders.localhost/api/orders
    ↓
[Host Network Interface - Port 80]
    ↓
[Traefik Container - Port 80]
    ↓ Reads labels from orderservice container
    ↓ Matches Host(`orders.localhost`)
    ↓ Finds service port: 3000
    ↓ 
[Web Network - Docker internal]
    ↓
[OrderService Container - Port 3000]
    ↓ Processes request
    ↓ Queries database via intercom network
    ↓
[PostgreSQL Container - Port 5432]
```

### Label Naming Convention
Traefik labels follow a hierarchical structure:
```
traefik.{protocol}.{type}.{name}.{property}={value}

Examples:
- traefik.http.routers.orderservice.rule=...
  └─ HTTP protocol, router type, name=orderservice, rule property

- traefik.http.services.orderservice.loadbalancer.server.port=...
  └─ HTTP protocol, service type, name=orderservice, loadbalancer config
```

## Why This Approach?

### 1. **Clean URLs**
- `http://orders.localhost` instead of `http://localhost:3000`
- Consistent with real-world domain usage
- Easy to add more services without port conflicts

### 2. **No Port Management**
- Don't need to track which service uses which port
- All services accessible through port 80
- Traefik handles internal routing

### 3. **Automatic Discovery**
- Traefik reads labels dynamically
- Changes take effect immediately when containers restart
- No manual Traefik configuration needed

### 4. **Scalability**
- Can scale orderservice: `docker-compose up --scale orderservice=3`
- Traefik automatically load balances between instances
- Zero configuration needed for load balancing

### 5. **Production-Ready**
- Same pattern works in production with real domains
- Easy to add HTTPS/TLS certificates later
- Can add middleware (authentication, rate limiting, etc.) via labels

## Testing the Configuration

Once containers are running, you can test:

1. **Check Traefik Dashboard:**
   - Visit: http://localhost:8080
   - Look for "orderservice" router in HTTP Routers section
   - Verify it shows as "enabled" with the Host rule

2. **Access the OrderService:**
   - Visit: http://orders.localhost
   - Should see the orderservice response
   - All API endpoints accessible: http://orders.localhost/api/orders

3. **Verify DNS Resolution:**
   - `*.localhost` domains resolve to 127.0.0.1 automatically on most systems
   - If not working, check `/etc/hosts` or local DNS settings

## Common Issues and Solutions

**Issue: "Gateway Timeout" or "Service Unavailable"**
- Check if orderservice container is running: `docker-compose ps`
- Verify orderservice is on the `web` network
- Check if orderservice is actually listening on port 3000

**Issue: "404 Page Not Found"**
- Verify the Host header is set correctly (use browser or curl with `-H "Host: orders.localhost"`)
- Check Traefik dashboard to see if the router is registered
- Ensure `traefik.enable=true` label is present

**Issue: Traefik doesn't see the container**
- Verify Docker socket is mounted in Traefik container
- Check if both containers share the `web` network
- Restart Traefik container: `docker-compose restart traefik`

---

# Part 3: SWS (Static Web Server) Configuration

## What Was Done
Configured SWS (Static Web Server) to serve the frontend static files (HTML, CSS, JavaScript) and made it accessible at `http://localhost` through Traefik routing.

## How It Was Implemented

### Docker Compose Service Configuration
Added a complete SWS service configuration in `docker-compose.yml`:

```yaml
sws:
  image: joseluisq/static-web-server:latest
  volumes:
    - ./frontend:/public
  environment:
    - SERVER_PORT=80
    - SERVER_ROOT=/public
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.sws.rule=Host(`localhost`)"
    - "traefik.http.routers.sws.entrypoints=web"
    - "traefik.http.services.sws.loadbalancer.server.port=80"
  networks:
    - web
```

### Breakdown of Each Configuration

#### 1. Image
```yaml
image: joseluisq/static-web-server:latest
```

**What:**
- Uses the official SWS (Static Web Server) Docker image
- A lightweight, fast, and modern static web server written in Rust

**Why SWS over nginx or Apache:**
- **Lightweight:** Much smaller image size (~2MB vs nginx ~130MB)
- **Simple:** Minimal configuration needed, works out of the box
- **Modern:** Built with modern web standards in mind
- **Performance:** High-performance static file serving
- **Environment-based config:** Easy to configure via environment variables

**Alternatives:**
- nginx - More features, heavier
- Apache - More features, much heavier
- Caddy - Similar lightweight approach

#### 2. Volume Mount
```yaml
volumes:
  - ./frontend:/public
```

**What:**
- Mounts the local `./frontend` directory into the container at `/public`
- `./frontend` is relative to the docker-compose.yml location
- `/public` is the path inside the container

**Why:**
- Makes local frontend files available to the SWS container
- Changes to frontend files are immediately visible (no rebuild needed)
- Separates frontend code from container image

**How it works:**
- Docker creates a bind mount between host and container
- Any file in `./frontend/` appears in `/public/` inside the container
- Files like `index.html`, CSS, JavaScript are served directly

**Development benefit:**
- Edit HTML/CSS/JS files locally
- Refresh browser to see changes
- No container rebuild required

#### 3. Environment Variables
```yaml
environment:
  - SERVER_PORT=80
  - SERVER_ROOT=/public
```

**SERVER_PORT=80**

**What:**
- Configures SWS to listen on port 80 inside the container

**Why:**
- Standard HTTP port
- Matches Traefik's expectation for web services
- Simpler configuration (no custom port needed)

**Note:**
- This is the container's internal port, not exposed to host
- Traefik routes to this port via Docker networking

**SERVER_ROOT=/public**

**What:**
- Sets the root directory from which SWS serves files
- Points to `/public` where we mounted the frontend files

**Why:**
- Tells SWS where to find the static files
- Must match the volume mount target path
- SWS will serve `index.html` from this directory

**How it works:**
- Request to `http://localhost/` → serves `/public/index.html`
- Request to `http://localhost/styles.css` → serves `/public/styles.css`
- Request to `http://localhost/js/app.js` → serves `/public/js/app.js`

#### 4. Traefik Labels
```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.sws.rule=Host(`localhost`)"
  - "traefik.http.routers.sws.entrypoints=web"
  - "traefik.http.services.sws.loadbalancer.server.port=80"
```

**traefik.enable=true**
- Same as orderservice - opts into Traefik routing

**traefik.http.routers.sws.rule=Host(\`localhost\`)**

**What:**
- Creates a router named "sws"
- Routes requests to `localhost` (no subdomain) to this service

**Why this is important:**
- Frontend at `http://localhost` (clean, simple URL)
- API at `http://orders.localhost` (clear separation)
- Different from orderservice which uses a subdomain

**Routing logic:**
```
http://localhost          → SWS (frontend)
http://orders.localhost   → OrderService (API)
http://api.localhost      → Could add more services
```

**traefik.http.routers.sws.entrypoints=web**
- Uses the same "web" entry point (port 80)
- Consistent with orderservice configuration

**traefik.http.services.sws.loadbalancer.server.port=80**

**What:**
- Tells Traefik that SWS listens on port 80 inside the container

**Why:**
- SWS is configured to listen on port 80 (via SERVER_PORT)
- Traefik needs to know this to forward traffic correctly
- Different from orderservice which uses port 3000

#### 5. Network Configuration
```yaml
networks:
  - web
```

**What:**
- Connects SWS only to the `web` network

**Why:**
- SWS only serves static files, doesn't need database access
- No need for `intercom` network (unlike orderservice)
- Security: Minimal network exposure

**Network comparison:**
```
traefik        → web only
orderservice   → web + intercom (needs database)
sws            → web only
order-postgres → intercom only (isolated)
```

## How the Complete System Works

### Request Flow for Frontend:
```
[Browser]
    ↓
    HTTP GET http://localhost/
    ↓
[Traefik on port 80]
    ↓ Matches Host(`localhost`)
    ↓ Routes to sws:80
    ↓
[SWS Container]
    ↓ Serves /public/index.html
    ↓
[Browser receives HTML]
```

### Request Flow for API from Frontend:
```
[Browser running frontend JS]
    ↓
    fetch('http://orders.localhost/api/orders')
    ↓
[Traefik on port 80]
    ↓ Matches Host(`orders.localhost`)
    ↓ Routes to orderservice:3000
    ↓
[OrderService Container]
    ↓ Queries database via intercom network
    ↓
[PostgreSQL]
```

## Complete Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                        Host System                       │
│                                                          │
│  Browser                                                 │
│    │                                                     │
│    ├─→ http://localhost ──────────┐                     │
│    └─→ http://orders.localhost ───┼─────────┐           │
│                                    │         │           │
└────────────────────────────────────┼─────────┼───────────┘
                                     ↓         ↓
                            ┌────────────────────────┐
                            │   Traefik (Port 80)    │
                            │   - Reverse Proxy      │
                            │   - Load Balancer      │
                            └──────────┬─────────────┘
                                       │
                        ┌──────────────┴──────────────┐
                        │      web network            │
                        │                             │
              ┌─────────┴─────────┐       ┌──────────┴─────────┐
              │       SWS         │       │   OrderService     │
              │   (Port 80)       │       │   (Port 3000)      │
              │ Serves Frontend   │       │   REST API         │
              └───────────────────┘       └──────────┬─────────┘
                                                     │
                                          ┌──────────┴─────────┐
                                          │  intercom network  │
                                          │                    │
                                          │  ┌──────────────┐  │
                                          └──┤  PostgreSQL  │  │
                                             │  (Port 5432) │  │
                                             └──────────────┘  │
                                                               │
                                             └─────────────────┘
```

## Why This Approach?

### 1. **Separation of Concerns**
- Static files served by specialized static server
- Dynamic API handled by application server
- Each service does one thing well

### 2. **Simple Frontend Development**
- Edit files locally, see changes immediately
- No build step needed for simple HTML/CSS/JS
- Clear separation from backend code

### 3. **Production-Like Setup**
- In production, frontend and backend are often separate
- CDN could replace SWS for static files
- Same routing patterns work with real domains

### 4. **Security**
- SWS has no database access
- Minimal attack surface (only serves static files)
- Can't execute arbitrary code

### 5. **Performance**
- SWS is optimized for static file serving
- Small, fast, efficient
- Can add caching headers easily

## Testing the Complete Setup

### 1. Build and Start Services
```bash
cd /home/stas24-2/fh3/SBD-AIS-Exercise/Exc_5/solution
docker-compose build
docker-compose up -d
```

### 2. Verify All Containers Are Running
```bash
docker-compose ps
```
Should show:
- traefik (running)
- orderservice (running)
- sws (running)
- order-postgres (running)

### 3. Check Traefik Dashboard
- Visit: http://localhost:8080
- Should see two routers:
  - `orderservice` → Host(`orders.localhost`)
  - `sws` → Host(`localhost`)

### 4. Test Frontend
- Visit: http://localhost
- Should see the frontend application
- Check browser console for any errors

### 5. Test API
- Visit: http://orders.localhost/api/orders
- Should see JSON response from orderservice
- Or test from frontend if it makes API calls

### 6. Test Frontend → API Communication
- If frontend has buttons/forms that call the API
- They should work because both are accessible via Traefik

## Common Issues and Solutions

### Issue: "This site can't be reached" for localhost
**Solution:**
- Check if containers are running: `docker-compose ps`
- Check if Traefik is running: `docker-compose logs traefik`
- Verify port 80 is not used by another service: `sudo lsof -i :80`

### Issue: Frontend loads but shows empty page
**Solution:**
- Check browser console for errors
- Verify files exist in `./frontend/` directory
- Check SWS logs: `docker-compose logs sws`
- Ensure `index.html` exists in frontend folder

### Issue: Frontend can't reach API
**Solution:**
- Check if API URL in frontend code uses `http://orders.localhost`
- Not `http://localhost:3000` (old port-based URL)
- CORS might be an issue - check orderservice CORS configuration

### Issue: Changes to frontend files not visible
**Solution:**
- Hard refresh browser: Ctrl+F5 (clear cache)
- Check if volume is mounted: `docker-compose exec sws ls /public`
- Verify you're editing files in the correct `./frontend/` directory

### Issue: "404 Not Found" on subdirectories
**Solution:**
- Check if files exist in the correct path
- SWS serves files relative to SERVER_ROOT
- Example: `/css/style.css` → `./frontend/css/style.css`

## Environment Variable Reference for SWS

Other useful SWS environment variables you can add:

```yaml
# Enable directory listing (useful for debugging)
SERVER_DIRECTORY_LISTING=true

# Enable CORS (if frontend and API have issues)
SERVER_CORS_ALLOW_ORIGINS=http://orders.localhost

# Set cache control headers
SERVER_CACHE_CONTROL_HEADERS=public, max-age=3600

# Enable compression
SERVER_COMPRESSION=true

# Custom error pages
SERVER_ERROR_PAGE_404=/404.html
SERVER_ERROR_PAGE_50x=/50x.html
```

## Next Steps and Enhancements

### Optional Improvements:
1. **Add HTTPS/TLS:**
   - Use Let's Encrypt with Traefik
   - Add HTTPS entry point
   - Redirect HTTP to HTTPS

2. **Add Middleware:**
   - Rate limiting via Traefik
   - Authentication for admin routes
   - Request logging

3. **Optimize Performance:**
   - Add caching headers in SWS
   - Enable compression
   - Use CDN in production

4. **Development Workflow:**
   - Add hot-reload for backend changes
   - Add file watchers for frontend builds
   - Add development vs production compose files

---

# Part 4: Building, Deploying, and Testing

## What Was Done
Built the Docker images, configured restart policies for resilience, and deployed all services to ensure they work together properly.

## Prerequisites

Before starting, ensure you have:
- Docker installed and running
- Docker Compose installed
- Go modules properly configured (go.mod and go.sum)
- Port 80, 5432, and 8080 available on your host machine

## Step-by-Step Deployment Process

### Step 1: Fix Go Dependencies

**Issue:** The initial build may fail due to missing `go.sum` entries.

**Solution:**
```bash
cd /home/stas24-2/fh3/SBD-AIS-Exercise/Exc_5/solution
go mod tidy
```

**What this does:**
- Downloads all required Go dependencies
- Updates `go.sum` with cryptographic checksums
- Ensures reproducible builds

**Why it's needed:**
- The `go.sum` file may be out of sync with `go.mod`
- `go mod tidy` resolves all dependencies and their versions
- Required before Docker build can succeed

### Step 2: Build Docker Images

```bash
cd /home/stas24-2/fh3/SBD-AIS-Exercise/Exc_5/solution
docker-compose build
```

**What happens:**
1. **OrderService build:**
   - Uses multi-stage Dockerfile
   - Stage 1: Compiles Go application using golang:1.25 image
   - Stage 2: Creates minimal runtime image using alpine
   - Runs `build-application.sh` script
   - Produces optimized binary (~10MB vs ~400MB)

2. **Other services (Traefik, SWS, PostgreSQL):**
   - Pull pre-built images from Docker Hub
   - No custom building required

**Expected output:**
```
[+] Building 18.4s (14/14) FINISHED
 => [builder 4/4] RUN sh /app/scripts/build-application.sh
 => [run 2/3] COPY --from=builder /app/ordersystem /app/ordersystem
 => exporting to image
 => => naming to docker.io/library/solution-orderservice:latest
```

### Step 3: Configure Restart Policies

Added resilience configurations to handle startup dependencies:

```yaml
orderservice:
  restart: on-failure
  depends_on:
    - order-postgres
```

**restart: on-failure**

**What:**
- Automatically restarts the container if it exits with a non-zero status code
- Docker will retry multiple times with exponential backoff

**Why:**
- OrderService may start before PostgreSQL is ready to accept connections
- First connection attempt fails, but retry succeeds once DB is ready
- Eliminates manual intervention for startup timing issues

**How it works:**
```
1. OrderService starts
2. PostgreSQL still initializing
3. Connection fails → Exit code 1
4. Docker waits 1 second
5. Restarts OrderService
6. PostgreSQL now ready
7. Connection succeeds → Service runs
```

**Alternatives:**
- `restart: always` - Restarts even on successful exit (not ideal)
- `restart: unless-stopped` - Restarts unless manually stopped
- `restart: "no"` - Never restart (default, not resilient)

**depends_on: [order-postgres]**

**What:**
- Defines startup order dependency
- Docker Compose starts `order-postgres` before `orderservice`

**Important limitation:**
- Only waits for container to start, NOT for application to be ready
- PostgreSQL container starts immediately, but initialization takes ~2-5 seconds
- This is why `restart: on-failure` is still needed

**Why both are needed:**
```
depends_on        → Ensures correct startup order
restart           → Handles timing issues within that order
```

**Enhanced alternative (optional):**
```yaml
depends_on:
  order-postgres:
    condition: service_healthy
```
This requires adding a healthcheck to PostgreSQL (covered in enhancements).

### Step 4: Start Services

**Option A: Foreground mode (recommended for testing)**
```bash
cd /home/stas24-2/fh3/SBD-AIS-Exercise/Exc_5/solution
docker-compose up
```

**Benefits:**
- See real-time logs from all containers
- Easy to spot errors immediately
- Stop with Ctrl+C

**Option B: Detached mode (background)**
```bash
docker-compose up -d
```

**Benefits:**
- Runs in background
- Terminal remains free
- View logs separately: `docker-compose logs -f`

### Step 5: Verify Container Status

```bash
docker-compose ps
```

**Expected output:**
```
NAME                        STATUS              PORTS
orderservice                Up                  
solution-order-postgres-1   Up                  0.0.0.0:5432->5432/tcp
solution-sws-1              Up                  
solution-traefik-1          Up                  0.0.0.0:80->80/tcp, 0.0.0.0:8080->8080/tcp
```

**All containers should show "Up" status**

### Step 6: Monitor Logs

Watch the startup sequence in the logs:

**PostgreSQL logs:**
```
order-postgres-1  | PostgreSQL init process complete; ready for start up.
order-postgres-1  | database system is ready to accept connections
```

**OrderService logs (first attempt - may fail):**
```
orderservice      | INFO Connecting to database
orderservice      | [error] failed to initialize database
orderservice      | dial tcp 172.19.0.2:5432: connect: connection refused
orderservice exited with code 1
```

**OrderService logs (retry - should succeed):**
```
orderservice      | INFO Connecting to database
orderservice      | INFO Database connection established
orderservice      | INFO Starting HTTP server on :3000
orderservice      | INFO Swagger docs available at /swagger/
```

**Traefik logs:**
```
traefik-1         | Configuration loaded from flags.
traefik-1         | Traefik version 3.5.2
```

**SWS logs:**
```
sws-1             | Server running at http://0.0.0.0:80
```

## Understanding the Startup Sequence

```
Time  Container         Status                    Notes
-------------------------------------------------------------------
T+0s  all               Starting                  Docker Compose begins
T+0s  traefik           Running                   No dependencies, starts immediately
T+0s  sws               Running                   No dependencies, starts immediately  
T+0s  order-postgres    Initializing              Starting PostgreSQL initialization
T+0s  orderservice      Waiting                   Waiting for order-postgres (depends_on)
T+2s  order-postgres    Initializing DB           Creating database "order"
T+2s  orderservice      Starting (1st attempt)    depends_on satisfied, container starts
T+2s  orderservice      Failed                    DB not accepting connections yet
T+3s  order-postgres    Ready                     "ready to accept connections"
T+4s  orderservice      Starting (2nd attempt)    restart: on-failure triggered
T+4s  orderservice      Running                   Successfully connected to DB
T+5s  ALL               Running                   System fully operational
```

## Testing the Complete Setup

### 1. Test Traefik Dashboard

```bash
# Open in browser or use curl
curl http://localhost:8080/api/overview
```

**Or visit in browser:** http://localhost:8080

**What to check:**
- Dashboard loads successfully
- "HTTP" section shows 2 routers:
  - `orderservice@docker` → Rule: Host(\`orders.localhost\`)
  - `sws@docker` → Rule: Host(\`localhost\`)
- Both routers show green "enabled" status
- "Services" section shows both services with status "UP"

### 2. Test Frontend (SWS)

```bash
curl http://localhost
```

**Expected:** HTML content from `frontend/index.html`

**Or visit in browser:** http://localhost

**What to check:**
- Page loads without errors
- No 404 errors in browser console
- Static assets (CSS, JS, images) load correctly
- Verify source: View Page Source should show your HTML

### 3. Test OrderService API

```bash
# Test the API endpoint
curl http://orders.localhost/api/orders
```

**Expected response:**
```json
[]
```
(Empty array initially, or existing orders if database has data)

**Or visit in browser:** http://orders.localhost/api/orders

**What to check:**
- JSON response received
- No connection errors
- Status code 200 OK

### 4. Test API Documentation (Swagger)

**Visit:** http://orders.localhost/swagger/index.html

**What to check:**
- Swagger UI loads
- API endpoints are documented
- Can test endpoints directly from Swagger UI

### 5. Test Frontend → API Communication

If your frontend makes API calls to the orderservice:

**Check browser console (F12):**
- No CORS errors
- API requests to `http://orders.localhost` succeed
- Network tab shows successful requests (200 status)

**Common frontend API call pattern:**
```javascript
fetch('http://orders.localhost/api/orders')
  .then(response => response.json())
  .then(data => console.log(data));
```

### 6. Test Database Connectivity

```bash
# Connect to PostgreSQL directly
docker-compose exec order-postgres psql -U docker -d order

# Inside psql, check tables
\dt

# Check if GORM migrations ran
SELECT * FROM drinks;
SELECT * FROM orders;
```

## Complete Testing Checklist

- [ ] All containers start successfully
- [ ] Traefik dashboard accessible at http://localhost:8080
- [ ] Frontend accessible at http://localhost
- [ ] API accessible at http://orders.localhost
- [ ] Swagger documentation at http://orders.localhost/swagger/index.html
- [ ] Frontend can make API calls successfully
- [ ] Database connections work (check orderservice logs)
- [ ] Changes to frontend files appear after refresh
- [ ] No CORS errors in browser console
- [ ] All routes in Traefik dashboard show as "enabled"
- [ ] OrderService automatically recovers if it exits

## Common Issues and Solutions

### Issue 1: Port Already in Use

**Error:**
```
Error starting userland proxy: listen tcp4 0.0.0.0:80: bind: address already in use
```

**Solution:**
```bash
# Find process using port 80
sudo lsof -i :80

# Stop the conflicting service (e.g., Apache)
sudo systemctl stop apache2

# Or change the port in docker-compose.yml
ports:
  - "8000:80"  # Use port 8000 instead
```

### Issue 2: OrderService Won't Start

**Symptoms:**
- Container keeps restarting
- Logs show database connection errors continuously

**Solutions:**

1. **Check PostgreSQL is running:**
   ```bash
   docker-compose ps order-postgres
   docker-compose logs order-postgres
   ```

2. **Verify network connectivity:**
   ```bash
   docker-compose exec orderservice ping order-postgres
   ```

3. **Check environment variables:**
   ```bash
   docker-compose exec orderservice env | grep POSTGRES
   ```

4. **Increase restart delay:**
   ```yaml
   restart: on-failure:5  # Max 5 retries
   ```

### Issue 3: "localhost" Domain Not Resolving

**Symptoms:**
- `curl http://localhost` works
- `curl http://orders.localhost` fails with "Could not resolve host"
- Frontend shows "Error loading menu"

**Solution:**

Most systems automatically resolve `*.localhost` to `127.0.0.1`, but **WSL and some systems do not**.

**For WSL (Ubuntu/Debian):**
```bash
# Add to /etc/hosts
echo "127.0.0.1 orders.localhost" | sudo tee -a /etc/hosts
```

**For Windows (REQUIRED when accessing from Windows browser):**

**⚠️ CRITICAL FOR WSL2 USERS:** You must use the WSL IP address, NOT 127.0.0.1!

1. **Find your WSL IP address:**
   - Open WSL terminal
   - Run: `hostname -I | awk '{print $1}'`
   - Note the IP (example: 172.31.100.169)

2. Open **Notepad as Administrator**
3. Open file: `C:\Windows\System32\drivers\etc\hosts`
4. Add these lines (replace `172.31.100.169` with YOUR WSL IP):
   ```
   172.31.100.169 orders.localhost
   172.31.100.169 localhost
   ```
   **DO NOT use 127.0.0.1** - it won't work with WSL2!

5. Save the file
6. Flush DNS cache:
   ```powershell
   ipconfig /flushdns
   ```
7. Verify:
   ```cmd
   ping orders.localhost
   ```
   Should respond from your WSL IP (e.g., 172.31.100.169)

**Important Notes:**
- WSL2 IP address can change when you restart WSL or your computer
- If it stops working, check the new IP with `wsl hostname -I` and update hosts file
- This is a WSL2 networking limitation, not a problem with our setup

**For macOS:**
```bash
# Add to /etc/hosts
echo "127.0.0.1 orders.localhost" | sudo tee -a /etc/hosts

# Flush DNS cache
sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
```

**For Linux:**
```bash
# Add to /etc/hosts
echo "127.0.0.1 orders.localhost" | sudo tee -a /etc/hosts
```

**Why this is needed:**
- The frontend JavaScript code runs in your **browser** (on Windows)
- It makes requests to `http://orders.localhost`
- Windows needs to know that `orders.localhost` points to `127.0.0.1`
- WSL's `/etc/hosts` only affects WSL, not Windows

### Issue 4: Frontend Can't Reach API (CORS Errors)

**Error in browser console:**
```
Access to fetch at 'http://orders.localhost/api/orders' from origin 'http://localhost' 
has been blocked by CORS policy
```

**Solution:**

Check orderservice CORS configuration in `main.go` or `rest/api.go`:

```go
// Ensure CORS allows localhost origin
cors.New(cors.Options{
    AllowedOrigins: []string{
        "http://localhost",
        "http://localhost:*",
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
})
```

### Issue 5: Changes to Go Code Not Reflected

**Solution:**
```bash
# Rebuild the image
docker-compose build orderservice

# Restart the service
docker-compose up -d --force-recreate orderservice
```

### Issue 6: Database Data Persists After docker-compose down

**This is expected** - volumes persist by design.

**To reset database:**
```bash
# Stop and remove volumes
docker-compose down -v

# Restart fresh
docker-compose up
```

### Issue 7: Frontend Shows "Error loading menu" (WSL Specific)

**Symptoms:**
- Frontend at http://localhost loads but shows "Error loading menu"
- Browser console shows: `Failed to fetch` or `net::ERR_NAME_NOT_RESOLVED`
- Order History and Total Orders sections are also empty

**Root Cause:**
This is **not a bug**, but a fundamental architectural challenge with host-based routing. The frontend JavaScript runs in your browser and needs to make API requests to `http://orders.localhost`. When accessing from a Windows browser while Docker runs in WSL, Windows doesn't automatically resolve `*.localhost` subdomains.

**Current Implementation (Host-Based Routing):**
```yaml
# docker-compose.yml configuration
orderservice:
  labels:
    - "traefik.http.routers.orderservice.rule=Host(`orders.localhost`)"

sws:
  labels:
    - "traefik.http.routers.sws.rule=Host(`localhost`)"
```

```javascript
// frontend/index.html - API calls use full URL
fetch("http://orders.localhost/api/menu")
fetch("http://orders.localhost/api/order/totalled")
fetch("http://orders.localhost/api/order/all")
```

**Why This Happens (Not WSL-Specific):**
1. Browser loads HTML from `http://localhost` (the `sws` service)
2. JavaScript executes and makes API calls to `http://orders.localhost`
3. This is a **cross-origin request** (different subdomains)
4. Requires DNS resolution for `orders.localhost`

**This affects ALL operating systems**, not just WSL:
- **Windows**: Doesn't auto-resolve `*.localhost` subdomains
- **macOS**: Some versions don't auto-resolve
- **Linux**: Depends on systemd-resolved configuration
- **WSL**: Additional complexity due to Windows-WSL networking

**Solution 1: Configure DNS Resolution (REQUIRED for current setup)**

**For WSL users accessing from Windows browser:**

1. **On Windows (your host OS):**
   - Open Notepad as Administrator
   - Open: `C:\Windows\System32\drivers\etc\hosts`
   - Add this line:
     ```
     127.0.0.1 orders.localhost
     ```
   - Save the file

2. **Flush DNS cache in Windows:**
   ```powershell
   # Run in PowerShell or CMD as Administrator
   ipconfig /flushdns
   ```

3. **Verify from Windows:**
   - Open Command Prompt
   - Run: `ping orders.localhost`
   - Should respond from 127.0.0.1

4. **Refresh your browser**
   - Hard refresh: Ctrl+Shift+R or Ctrl+F5
   - Check browser console (F12) - no more errors
   - Menu should now load

**For Linux/macOS users:**
```bash
# Add to /etc/hosts
echo "127.0.0.1 orders.localhost" | sudo tee -a /etc/hosts
```

**Pros of Host-Based Routing:**
- ✅ Clear separation between services (different domains)
- ✅ Production-like setup (mirrors real-world architecture)
- ✅ Easy to add more services with different subdomains
- ✅ Works with CORS once properly configured

**Cons of Host-Based Routing:**
- ❌ Requires DNS configuration (hosts file modification)
- ❌ Need to configure CORS headers in API
- ❌ More complex for development environments
- ❌ DNS configuration can break after system updates

**Solution 2: Path-Based Routing (Alternative Architecture)**

A more elegant solution that eliminates DNS issues entirely. See **Part 5** below for complete implementation guide.

## Useful Commands

### Rebuild Docker Containers

**When do you need to rebuild?**
- After changing Go source code (main.go, rest/api.go, etc.)
- After modifying Dockerfile
- After updating dependencies (go.mod, go.sum)

**Rebuild Commands:**

```bash
# Method 1: Rebuild only orderservice (fastest)
docker-compose build orderservice

# Method 2: Rebuild without using cache (clean build)
docker-compose build --no-cache orderservice

# Method 3: Rebuild and immediately restart the service
docker-compose up -d --build orderservice

# Method 4: Rebuild all services that have a build directive
docker-compose build

# Method 5: Complete rebuild - stop everything, rebuild, restart
docker-compose down
docker-compose build
docker-compose up -d
```

**What gets rebuilt:**
- Only services with `build: .` directive get rebuilt (orderservice)
- Services using pre-built images (traefik, sws, postgres) are just pulled, not built

**No rebuild needed for:**
- Frontend changes (HTML/CSS/JS) - just refresh browser
- docker-compose.yml changes - just run `docker-compose up -d`
- Environment variable changes - run `docker-compose up -d --force-recreate`

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f orderservice

# Last 100 lines
docker-compose logs --tail=100 orderservice
```

### Restart specific service
```bash
docker-compose restart orderservice
```

### Force recreate without rebuilding
```bash
docker-compose up -d --force-recreate orderservice
```

### Execute commands in containers
```bash
# Shell into orderservice
docker-compose exec orderservice sh

# Run psql
docker-compose exec order-postgres psql -U docker -d order
```

### Check network connectivity
```bash
# From orderservice to database
docker-compose exec orderservice ping order-postgres

# From orderservice to traefik
docker-compose exec orderservice ping traefik
```

### View container resource usage
```bash
docker stats
```

### Clean up everything
```bash
# Stop and remove containers, networks
docker-compose down

# Also remove volumes (deletes database data)
docker-compose down -v

# Remove images too
docker-compose down --rmi all -v
```

## Performance Optimization Tips

### 1. Use Healthchecks
Add to `docker-compose.yml`:
```yaml
order-postgres:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U docker"]
    interval: 5s
    timeout: 5s
    retries: 5
```

### 2. Optimize OrderService Startup
Add connection retry logic with backoff in Go code.

### 3. Add Resource Limits
```yaml
orderservice:
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
```

### 4. Enable Compression in SWS
```yaml
sws:
  environment:
    - SERVER_COMPRESSION=true
    - SERVER_COMPRESSION_LEVEL=6
```

## Summary

You now have a complete microservices setup with:
- ✅ **Traefik** - Reverse proxy routing traffic based on domains
- ✅ **SWS** - Serving static frontend files at http://localhost
- ✅ **OrderService** - REST API accessible at http://orders.localhost
- ✅ **PostgreSQL** - Database isolated on internal network
- ✅ **Network Isolation** - Frontend and backend separated from database
- ✅ **Automatic Restart** - Services recover from transient failures
- ✅ **Startup Dependencies** - Correct initialization order
- ✅ **Production-Ready Pattern** - Easy to migrate to real domains

The architecture is scalable, secure, resilient, and follows modern microservices best practices!

---

# Part 5: Alternative Routing Architecture - Path-Based Routing

## Overview

This section describes an alternative to the host-based routing currently implemented. Path-based routing eliminates DNS configuration issues and simplifies frontend-backend communication.

## Current Architecture vs Alternative

### Current Implementation (Host-Based Routing)
```
Frontend:  http://localhost       → SWS
API:       http://orders.localhost → OrderService
```

**Requires:**
- DNS configuration (hosts file)
- CORS configuration
- Full URLs in frontend code

### Alternative Implementation (Path-Based Routing)
```
Frontend:  http://localhost/      → SWS
API:       http://localhost/api/* → OrderService
```

**Requires:**
- No DNS configuration
- No CORS configuration (same origin)
- Relative URLs in frontend code

## How Path-Based Routing Works

### Concept
Instead of routing based on domain/subdomain (Host header), Traefik routes based on URL path:

```
Request Path                 → Routed To
─────────────────────────────────────────────
http://localhost/            → SWS (frontend)
http://localhost/index.html  → SWS (frontend)
http://localhost/api/menu    → OrderService
http://localhost/api/orders  → OrderService
http://localhost/openapi/*   → OrderService
```

### How Traefik Decides

Traefik evaluates rules with **priority**:
1. Higher priority rules are checked first
2. First matching rule wins
3. Request is routed to the corresponding service

```
Priority 2: PathPrefix(`/api`) → OrderService
Priority 1: Host(`localhost`)  → SWS (catch-all)
```

## Implementation Guide

### Step 1: Update docker-compose.yml

Replace the Traefik labels in your `docker-compose.yml`:

```yaml
services:
  traefik:
    # ... existing traefik configuration unchanged ...

  orderservice:
    container_name: orderservice
    build: .
    command: [ "/app/ordersystem" ]
    restart: on-failure
    depends_on:
      - order-postgres
    labels:
      # Enable Traefik routing for this container
      - "traefik.enable=true"
      
      # Route based on path prefix, not hostname
      - "traefik.http.routers.orderservice.rule=Host(`localhost`) && PathPrefix(`/api`)"
      
      # Use web entry point (port 80)
      - "traefik.http.routers.orderservice.entrypoints=web"
      
      # Higher priority to match before catch-all SWS rule
      - "traefik.http.routers.orderservice.priority=2"
      
      # Specify the port where the service is listening
      - "traefik.http.services.orderservice.loadbalancer.server.port=3000"
      
      # Use the web network for routing
      - "traefik.docker.network=web"
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER=docker
      - POSTGRES_PASSWORD=docker
      - POSTGRES_TCP_PORT=5432
      - DB_HOST=order-postgres
    networks:
      - intercom
      - web

  sws:
    image: joseluisq/static-web-server:latest
    volumes:
      - ./frontend:/public
    environment:
      - SERVER_PORT=80
      - SERVER_ROOT=/public
    labels:
      # Enable Traefik routing for this container
      - "traefik.enable=true"
      
      # Catch-all rule for localhost (no path restriction)
      - "traefik.http.routers.sws.rule=Host(`localhost`)"
      
      # Use web entry point (port 80)
      - "traefik.http.routers.sws.entrypoints=web"
      
      # Lower priority - catches everything not matched by orderservice
      - "traefik.http.routers.sws.priority=1"
      
      # Specify the port where SWS is listening
      - "traefik.http.services.sws.loadbalancer.server.port=80"
    networks:
      - web

  order-postgres:
    # ... unchanged ...
```

### Step 2: Update frontend/index.html

Change the API URLs from absolute to relative:

**Before (Host-Based Routing):**
```javascript
fetch("http://orders.localhost/api/menu")
fetch("http://orders.localhost/api/order/totalled")
fetch("http://orders.localhost/api/order/all")
```

**After (Path-Based Routing):**
```javascript
fetch("/api/menu")
fetch("/api/order/totalled")
fetch("/api/order/all")
```

**Complete example:**
```javascript
// Fetch and display the drink menu
// Using relative URL - works through Traefik path-based routing
fetch("/api/menu")
    .then(res => res.json())
    .then(drinks => {
        menuList.innerHTML = "";
        drinks.forEach(drink => {
            drinkMap.set(drink.id, drink);
            const li = document.createElement("li");
            li.textContent = `${drink.name} - $${drink.price.toFixed(2)} (${drink.description})`;
            menuList.appendChild(li);
        });

        // After loading drinks, load orders
        loadOrderTotalled();
        loadOrders();
    })
    .catch(err => {
        console.error("Failed to load menu:", err);
        menuList.innerHTML = "<li>Error loading menu.</li>";
    });

function loadOrderTotalled() {
    fetch("/api/order/totalled")
        .then(res => res.json())
        .then(orders => {
            // ... existing code ...
        })
        .catch(err => {
            console.error("Failed to load orders:", err);
            totalledChartDiv.textContent = "Error loading orders.";
        });
}

function loadOrders() {
    fetch("/api/order/all")
        .then(res => res.json())
        .then(orders => {
            // ... existing code ...
        })
        .catch(err => {
            console.error("Failed to load orders:", err);
            ordersChartDiv.textContent = "Error loading orders.";
        });
}
```

### Step 3: Deploy and Test

```bash
# Stop existing containers
docker-compose down

# Start with new configuration
docker-compose up -d

# Wait for services to start
sleep 5

# Test the setup
curl http://localhost/           # Should return HTML (frontend)
curl http://localhost/api/menu   # Should return JSON (API)
```

## Understanding Priority in Traefik

### Why Priority Matters

Without priority, Traefik might match the wrong rule:

```
Request: http://localhost/api/menu

Without priority:
  ❌ Matches Host(`localhost`) first → Routes to SWS → 404 error

With priority:
  ✅ Checks priority=2 first: Host(`localhost`) && PathPrefix(`/api`) → Matches!
  ✅ Routes to OrderService → Success
```

### How Priority Works

```yaml
# Priority 2 - Checked first
orderservice:
  - "traefik.http.routers.orderservice.rule=Host(`localhost`) && PathPrefix(`/api`)"
  - "traefik.http.routers.orderservice.priority=2"

# Priority 1 - Checked second (default is 0)
sws:
  - "traefik.http.routers.sws.rule=Host(`localhost`)"
  - "traefik.http.routers.sws.priority=1"
```

**Rule evaluation order:**
1. Check orderservice rule (priority 2)
   - If path starts with `/api` → Route to orderservice
2. Check sws rule (priority 1)
   - Matches any localhost request → Route to sws

### PathPrefix Explanation

`PathPrefix(`/api`)` matches:
- ✅ `/api`
- ✅ `/api/`
- ✅ `/api/menu`
- ✅ `/api/orders/123`
- ✅ `/api/anything/deeply/nested`

Does NOT match:
- ❌ `/`
- ❌ `/index.html`
- ❌ `/styles.css`
- ❌ `/apidocs` (doesn't start with `/api/` exactly)

## Advanced PathPrefix Patterns

You can route multiple path prefixes to the same service:

```yaml
# Route both /api and /openapi to orderservice
- "traefik.http.routers.orderservice.rule=Host(`localhost`) && (PathPrefix(`/api`) || PathPrefix(`/openapi`))"
```

Or create separate routers:

```yaml
# API router
- "traefik.http.routers.orderservice-api.rule=Host(`localhost`) && PathPrefix(`/api`)"
- "traefik.http.routers.orderservice-api.priority=2"
- "traefik.http.routers.orderservice-api.service=orderservice"

# OpenAPI/Swagger router
- "traefik.http.routers.orderservice-docs.rule=Host(`localhost`) && PathPrefix(`/openapi`)"
- "traefik.http.routers.orderservice-docs.priority=2"
- "traefik.http.routers.orderservice-docs.service=orderservice"

# Service definition (shared by both routers)
- "traefik.http.services.orderservice.loadbalancer.server.port=3000"
```

## Request Flow Diagram

### Path-Based Routing Flow

```
┌─────────────────────────────────────────────────────────┐
│                    Browser                              │
│                                                         │
│  User visits: http://localhost/                        │
│  JavaScript loads and makes:                           │
│    fetch('/api/menu')                                  │
│    fetch('/api/order/totalled')                        │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓ All requests to localhost:80
         ┌────────────────────┐
         │   Traefik          │
         │   Entry Point:web  │
         └────────┬───────────┘
                  │
                  ├─ Request: / (GET index.html)
                  │    ↓ Check rules by priority
                  │    ↓ Priority 2: /api? No
                  │    ↓ Priority 1: localhost? Yes
                  │    ↓
                  │  ┌──────────────┐
                  │  │     SWS      │
                  │  │  Port: 80    │
                  │  │ /public/     │
                  │  └──────┬───────┘
                  │         │
                  │         └─→ Returns index.html
                  │
                  └─ Request: /api/menu
                       ↓ Check rules by priority
                       ↓ Priority 2: localhost && /api? Yes
                       ↓
                     ┌──────────────────┐
                     │  OrderService    │
                     │   Port: 3000     │
                     │   REST API       │
                     └────────┬─────────┘
                              │
                              ├─→ Queries PostgreSQL
                              │   (via intercom network)
                              │
                              └─→ Returns JSON
```

## Comparison: Host-Based vs Path-Based

| Aspect | Host-Based (Current) | Path-Based (Alternative) |
|--------|---------------------|--------------------------|
| **DNS Setup** | ❌ Required (hosts file) | ✅ Not needed |
| **CORS** | ❌ Required (cross-origin) | ✅ Not needed (same origin) |
| **Frontend Code** | Full URLs: `http://orders.localhost/api/menu` | Relative URLs: `/api/menu` |
| **Production Mapping** | Easy: subdomain → service | Slightly harder: path routing |
| **Adding Services** | New subdomain per service | New path prefix per service |
| **WSL Compatibility** | ❌ Needs Windows hosts file | ✅ Works out of the box |
| **Complexity** | Higher (DNS + CORS) | Lower (just Traefik config) |
| **Security** | Better isolation via subdomains | Same origin (less isolation) |
| **URL Clarity** | `orders.localhost` (very clear) | `localhost/api` (clear enough) |
| **Scalability** | Excellent (microservices pattern) | Good (monolith with modules) |
| **Mobile Testing** | Harder (subdomain resolution) | Easier (single domain) |

## When to Use Each Approach

### Use Host-Based Routing When:
- ✅ Deploying to production with real domains (api.example.com, app.example.com)
- ✅ Need strong service isolation
- ✅ Multiple teams manage different services
- ✅ Services have completely different security requirements
- ✅ You're okay with DNS configuration

### Use Path-Based Routing When:
- ✅ Development environment (especially with WSL)
- ✅ Want simplest possible setup
- ✅ Services are tightly coupled (frontend + backend)
- ✅ No DNS configuration possible (restricted environments)
- ✅ Testing on mobile devices via local network
- ✅ Prototyping or learning

## Migration Between Approaches

### From Host-Based → Path-Based

**Changes required:**
1. Update `docker-compose.yml` labels (add PathPrefix, priority)
2. Update `frontend/index.html` (absolute → relative URLs)
3. No CORS configuration needed (can remove if present)
4. No hosts file entries needed (can remove)

**Benefits:**
- Simpler development workflow
- No DNS issues
- Works immediately in WSL

### From Path-Based → Host-Based

**Changes required:**
1. Update `docker-compose.yml` labels (use Host rules only)
2. Update `frontend/index.html` (relative → absolute URLs)
3. Add CORS configuration to API
4. Configure DNS (hosts file or real DNS)

**Benefits:**
- Better separation of concerns
- More production-like
- Clearer service boundaries

## Testing Path-Based Routing

### 1. Check Traefik Dashboard

Visit: http://localhost:8080

**Look for:**
- Two HTTP routers:
  - `orderservice@docker` - Priority: 2, Rule: `Host(\`localhost\`) && PathPrefix(\`/api\`)`
  - `sws@docker` - Priority: 1, Rule: `Host(\`localhost\`)`

### 2. Test Routes

```bash
# Frontend (should return HTML)
curl http://localhost/

# API menu (should return JSON)
curl http://localhost/api/menu

# API orders (should return JSON array)
curl http://localhost/api/orders

# Swagger docs (if configured)
curl http://localhost/openapi/index.html
```

### 3. Browser Testing

1. Open: http://localhost/
2. Open browser console (F12)
3. Network tab should show:
   - `/ ` → 200 OK (HTML from SWS)
   - `/api/menu` → 200 OK (JSON from OrderService)
   - `/api/order/totalled` → 200 OK (JSON from OrderService)
   - `/api/order/all` → 200 OK (JSON from OrderService)
4. **No CORS errors** (same origin!)

## Troubleshooting Path-Based Routing

### Issue: API requests return HTML instead of JSON

**Cause:** Traefik is routing to SWS instead of OrderService

**Solution:**
- Check priority is set correctly (orderservice must be higher)
- Verify PathPrefix syntax: `PathPrefix(\`/api\`)`
- Restart Traefik: `docker-compose restart traefik`

### Issue: API requests return 404

**Cause:** OrderService isn't handling the path correctly

**Check:**
```bash
# Direct test (bypass Traefik)
docker-compose exec orderservice wget -O- http://localhost:3000/api/menu
```

If this works, issue is in Traefik routing.
If this fails, issue is in OrderService API configuration.

### Issue: Priority doesn't seem to work

**Solution:**
- Ensure priority values are integers (no quotes)
- Higher number = higher priority
- Restart Traefik to apply changes
- Check Traefik logs: `docker-compose logs traefik`

## Summary

**Current Implementation (Host-Based Routing):**
- Uses subdomains: `localhost` and `orders.localhost`
- Requires DNS configuration (hosts file)
- Frontend uses absolute URLs: `http://orders.localhost/api/menu`
- Requires CORS configuration
- More production-like architecture

**Alternative Implementation (Path-Based Routing):**
- Uses paths: `localhost/` and `localhost/api/*`
- No DNS configuration needed
- Frontend uses relative URLs: `/api/menu`
- No CORS configuration needed
- Simpler development setup

**Both approaches are valid** - choose based on your requirements:
- **Learning/Development**: Path-based is simpler
- **Production/Real-world**: Host-based is more realistic

**Recommendation for WSL users:** Consider using path-based routing to avoid the complexity of managing Windows hosts file entries, especially since WSL IP addresses can change.

````
