# Exercise 8 Solution - gRPC Order Service

## Overview
This solution implements a gRPC-based drink ordering system with a server and client in Go.

---

## File Structure & Implementation

### 1. `pb/orders.proto` - Protocol Buffer Definition
**Purpose:** Defines the data structures and service contract for gRPC communication.

- **Messages:**
  - `Drink` - represents a drink item (id, name, price, description)
  - `DrinkList` - container for multiple drinks
  - `OrderRequest` - request to order a drink (drink_id, quantity)
  - `OrderItem` - ordered drink with quantity
  - `OrderList` - container for all orders (the bill)

- **Service:** `OrderService` with 3 RPC methods:
  - `GetDrinks` - returns available drinks
  - `OrderDrink` - places an order
  - `GetOrders` - returns all orders

---

### 2. `server/grpc_server.go` - gRPC Server Implementation
**Purpose:** Implements the `OrderServiceServer` interface.

- **Storage:** In-memory maps for drinks and orders (no database needed)
- **Thread Safety:** Uses mutex for concurrent order access
- **Prepopulated Drinks:** Spritzer (€2), Beer (€3), Coffee (free)
- **Embeds:** `UnimplementedOrderServiceServer` for forward compatibility

---

### 3. `client/grpc_client.go` - gRPC Client Implementation
**Purpose:** Demonstrates client interaction with the order service.

- Connects to server on port 4000
- Executes workflow: list drinks → order 2 each → order 6 more → get bill
- Uses insecure credentials (local development)

---

### 4. `main.go` - Application Entry Point
**Purpose:** Starts server and runs client.

- Starts gRPC server in a goroutine (concurrent execution)
- Waits 1 second for server startup
- Creates client and runs the ordering workflow

---

## Installation & Setup

### Install Protocol Buffer Compiler
```bash
# macOS (Homebrew)
brew install protobuf protoc-gen-go protoc-gen-go-grpc

# Or download from: https://protobuf.dev/installation/
```

### Generate Go Code from Proto
```bash
cd Exc_8/solution
chmod +x generate_pb.sh
./generate_pb.sh
```

This generates:
- `pb/orders.pb.go` - message types
- `pb/orders_grpc.pb.go` - service client/server interfaces

---

## Running & Testing

### Run the Application
```bash
cd Exc_8/solution
go run .
```

### Expected Output
```
Requesting drinks 🍹🍺☕
Available drinks:
    > id:1  name:"Spritzer"  price:2  description:"Wine with soda"
    > id:2  name:"Beer"  price:3  description:"Hagenberger Gold"
    > id:3  name:"Coffee"  price:0  description:"Mifare isn't that secure"
Ordering drinks 👨‍🍳⏱️🍻🍻
    > Ordering: 2 x Spritzer
    > Ordering: 2 x Beer
    > Ordering: 2 x Coffee
Ordering another round of drinks 👨‍🍳⏱️🍻🍻
    > Ordering: 6 x Spritzer
    > Ordering: 6 x Beer
    > Ordering: 6 x Coffee
Getting the bill 💹💹💹
    > Total: 8 x Spritzer
    > Total: 8 x Beer
    > Total: 8 x Coffee
Orders complete!
```

---

## Key Concepts

| Concept | Description |
|---------|-------------|
| **gRPC** | High-performance RPC framework using HTTP/2 |
| **Protocol Buffers** | Language-neutral serialization format |
| **Unary RPC** | Simple request-response pattern used here |
| **Forward Compatibility** | Embedding `Unimplemented*` struct handles future proto changes |
