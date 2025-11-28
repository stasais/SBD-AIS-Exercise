# Implementation Notes

## Target Architecture (2 Nodes)

```
              2 Node Swarm Setup              
                                                                                    
┌──────────────┐   ┌──────────────┐
│    Node 1    │   │    Node 2    │
│   stasdell   │   │docker-desktop│
├──────────────┤   ├──────────────┤
│  Manager     │   │   Worker     │
├──────────────┤   ├──────────────┤
│   Traefik    │   │  PostgresDb  │
│     SWS      │   │    Minio     │
│ Orderservice │   │     SWS      │
│              │   │ Orderservice │
└──────────────┘   └──────────────┘
```

---

## Network Configuration

### Check Local IP Address

Command to find your local network IP addresses:
```bash
ip addr show | grep -E "inet " 
```

for mac
```bash
ifconfig | grep "inet "  | grep -v "
```

### Current IP Addresses

| Interface | IP Address | Description |
|-----------|------------|-------------|
| `wlp0s20f3` | **192.168.1.64** | Main WiFi network IP (Linux) |
| `tailscale0` | **100.67.76.54** | Tailscale VPN (Linux) |
| `en0` | **192.168.1.152** | Main WiFi network IP (Mac) |
| `utun` | **100.89.0.126** | Tailscale VPN (Mac) |

---

## Swarm Setup

### 1. Init Swarm (Manager - Linux)
```bash
docker swarm init --advertise-addr 192.168.1.64
```

### 2. Join Swarm (Worker - Mac)
```bash
docker swarm join --token SWMTKN-1-4ssd0kvmf9ertxqubfmj43xz2vf7wqqj8iiwluu3175h1qutzl-bz5nep5drbrcamjiqvbnl6etq 192.168.1.64:2377
docker swarm join --token SWMTKN-1-5es43yjmaf9n4w834xodz1qmnaaspn244hrplqg9svkz34enb4-dibeon9740rhglmun0nf7sqra 100.67.76.54:2377
```

### 3. Verify Nodes
```bash
docker node ls
```

### 4. Deploy Stack
```bash
docker stack deploy --compose-file docker-compose.yml sbd
```

---

## Initial Setup Analysis

### Docker Compose Structure

#### Networks (✅ Configured)
| Network | Driver | Purpose |
|---------|--------|---------|
| `web` | overlay | External traffic - Traefik routing |
| `intercom` | overlay | Internal service communication |

#### Volumes (✅ Configured)
| Volume | Purpose |
|--------|---------|
| `order_pg_vol` | PostgreSQL data persistence |
| `minio_vol` | Minio S3 data persistence |

#### Secrets (✅ Files Created)
| Secret | File Location | Current Value |
|--------|---------------|---------------|
| `postgres_user` | `docker/postgres_user_secret` | `docker` |
| `postgres_password` | `docker/postgres_password_secret` | `docker` |
| `s3_user` | `docker/s3_user_secret` | `root` |
| `s3_password` | `docker/s3_password_secret` | `verysecret` |

---

### Services Status

#### 1. Traefik (Reverse Proxy)
| Setting | Status | Current Value |
|---------|--------|---------------|
| Image | ✅ | `traefik:v3.6.1` |
| Port | ✅ | 80 (ingress mode) |
| Swarm Provider | ✅ | `providers.swarm.endpoint` configured |
| Network | ✅ | `web` |
| Dashboard | ⚠️ TODO | Needs labels for `/dashboard` |
| Deploy constraints | ❌ TODO | Should be `node.role == manager` |

#### 2. Frontend (SWS)
| Setting | Status | Current Value |
|---------|--------|---------------|
| Image | ✅ | `ghcr.io/ddibiasi/sbd-ais-exercise/frontend:1.0` |
| Network | ✅ | `web` |
| Deploy mode | ❌ TODO | Should be `mode: global` |
| Traefik labels | ❌ TODO | Needs routing to `http://localhost` |

#### 3. Orderservice (API)
| Setting | Status | Current Value |
|---------|--------|---------------|
| Image | ✅ | `ghcr.io/ddibiasi/sbd-ais-exercise/orderservice:1.0` |
| Networks | ✅ | `web`, `intercom` |
| Restart policy | ✅ | `condition: any` |
| Deploy mode | ❌ TODO | Should be `mode: global` |
| Traefik labels | ❌ TODO | Needs routing to `http://orders.localhost` |
| Secrets binding | ❌ TODO | `*_FILE` env vars are empty |

#### 4. PostgreSQL (Database)
| Setting | Status | Current Value |
|---------|--------|---------------|
| Image | ✅ | `postgres:18` |
| Volume | ✅ | `order_pg_vol` |
| Network | ✅ | `intercom` |
| Port | ✅ | `5555` |
| Deploy constraints | ❌ TODO | Should bind to specific worker node |
| Secrets binding | ❌ TODO | `*_FILE` env vars are empty |

#### 5. Minio (S3 Storage)
| Setting | Status | Current Value |
|---------|--------|---------------|
| Image | ✅ | `minio/minio:latest` |
| Volume | ✅ | `minio_vol` |
| Network | ✅ | `intercom` |
| Port | ✅ | `8500` |
| Deploy constraints | ❌ TODO | Should bind to specific worker node |
| Secrets binding | ❌ TODO | `*_FILE` env vars are empty |

---

### TODO Summary

| Priority | Task | Services Affected | Status |
|----------|------|-------------------|--------|
| 1 | Add deploy constraint `node.role == manager` | Traefik | ✅ Done |
| 2 | Add `mode: global` deployment | Frontend, Orderservice | ✅ Done |
| 3 | Add node hostname constraints | PostgreSQL, Minio | ✅ Done (on manager due to Mac arm64 issues) |
| 4 | Wire secrets to `*_FILE` env vars | Orderservice, PostgreSQL, Minio | ✅ Done |
| 5 | Add Traefik labels under `deploy:` | Frontend, Orderservice, Traefik dashboard | ✅ Done |

---

## API Endpoints

| Endpoint | Command |
|----------|---------|
| Menu | `curl http://orders.192.168.1.64.nip.io/api/menu` |
| All Orders | `curl http://orders.192.168.1.64.nip.io/api/order/all` |
| Totalled Orders | `curl http://orders.192.168.1.64.nip.io/api/order/totalled` |
| Place Order | `curl -X POST -H "Content-Type: application/json" -d '{"drink_id":1,"amount":2}' http://orders.192.168.1.64.nip.io/api/order` |
| Get Receipt | `curl http://orders.192.168.1.64.nip.io/api/receipt/1` |

---

## Access URLs

| Service | URL |
|---------|-----|
| Frontend | `http://192.168.1.64/` |
| API (via nip.io) | `http://orders.192.168.1.64.nip.io/api/menu` |
| Traefik Dashboard | `http://192.168.1.64/dashboard/` |

---

## Current Setup (Tailscale VPN)

Using Tailscale for reliable cross-node networking. Swarm initialized with `--advertise-addr 100.67.76.54`.

### Tailscale API Endpoints

| Endpoint | Command |
|----------|---------|
| Menu | `curl http://orders.100.67.76.54.nip.io/api/menu` |
| All Orders | `curl http://orders.100.67.76.54.nip.io/api/order/all` |
| Totalled Orders | `curl http://orders.100.67.76.54.nip.io/api/order/totalled` |
| Place Order | `curl -X POST -H "Content-Type: application/json" -d '{"drink_id":1,"amount":2}' http://orders.100.67.76.54.nip.io/api/order` |
| Get Receipt | `curl http://orders.100.67.76.54.nip.io/api/receipt/1` |

### Tailscale Access URLs

| Service | URL |
|---------|-----|
| Frontend | `http://100.67.76.54/` |
| API (via nip.io) | `http://orders.100.67.76.54.nip.io/api/menu` |
| API (path-based) | `http://100.67.76.54/api/menu` |
| Traefik Dashboard | `http://100.67.76.54/dashboard/` |

### Node IPs (Tailscale)

| Node | Tailscale IP |
|------|--------------|
| stasdell (Manager) | `100.67.76.54` |
| docker-desktop (Worker) | `100.89.0.126` |



