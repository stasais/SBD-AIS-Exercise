package main

import (
	// Import server package to start gRPC server
	"exc8/server"
	// Import client package to run gRPC client
	"exc8/client"
	// Import log for error logging
	"log"
	"time"
)

// ============================================================================
// MAIN FUNCTION
// Entry point that starts server and runs client
// ============================================================================
func main() {
	// Start server in a goroutine so it runs concurrently
	// This allows the client to connect while server is running
	go func() {
		// todo start server
		// Call StartGrpcServer to initialize and run the gRPC server
		// Server listens on port 4000 as defined in server package
		err := server.StartGrpcServer()
		if err != nil {
			// Log fatal error if server fails to start
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
	// Wait for server to start up before connecting client
	// 1 second delay ensures server is ready to accept connections
	time.Sleep(1 * time.Second)
	// todo start client
	// Create new gRPC client instance
	grpcClient, err := client.NewGrpcClient()
	if err != nil {
		// Log fatal error if client connection fails
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	// Run the client workflow (list drinks, order, get bill)
	err = grpcClient.Run()
	if err != nil {
		// Log fatal error if client workflow fails
		log.Fatalf("Client run failed: %v", err)
	}
	println("Orders complete!")
}
