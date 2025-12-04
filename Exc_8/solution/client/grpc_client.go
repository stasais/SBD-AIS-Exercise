package client

import (
	"context"
	"exc8/pb"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ============================================================================
// GRPC CLIENT STRUCT
// Client wrapper for interacting with the OrderService
// ============================================================================
type GrpcClient struct {
	// The gRPC client stub generated from protobuf
	client pb.OrderServiceClient
}

// ============================================================================
// NEW GRPC CLIENT CONSTRUCTOR
// Creates a new client connection to the gRPC server
// ============================================================================
func NewGrpcClient() (*GrpcClient, error) {
	// Create new gRPC client connection to server on port 4000
	// Using insecure credentials for local development (no TLS)
	conn, err := grpc.NewClient(":4000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// Return error if connection fails
		return nil, err
	}
	// Create the OrderService client from the connection
	client := pb.NewOrderServiceClient(conn)
	// Return wrapped client
	return &GrpcClient{client: client}, nil
}

// ============================================================================
// RUN CLIENT WORKFLOW
// Executes the complete order workflow as specified in README
// ============================================================================
func (c *GrpcClient) Run() error {
	// todo
	// 1. List drinks
	// 2. Order a few drinks
	// 3. Order more drinks
	// 4. Get order total
	//
	// print responses after each call
	
	// Create background context for all RPC calls
	ctx := context.Background()
	
	// ========================================================================
	// STEP 1: LIST DRINKS
	// Request and display all available drinks
	// ========================================================================
	fmt.Println("Requesting drinks 🍹🍺☕")
	// Call GetDrinks RPC with empty request
	drinks, err := c.client.GetDrinks(ctx, &emptypb.Empty{})
	if err != nil {
		// Return error if RPC fails
		return fmt.Errorf("failed to get drinks: %w", err)
	}
	fmt.Println("Available drinks:")
	// Iterate and print each available drink
	for _, drink := range drinks.Drinks {
		fmt.Printf("\t> id:%d  name:%q  price:%d  description:%q\n", 
			drink.Id, drink.Name, drink.Price, drink.Description)
	}
	
	// ========================================================================
	// STEP 2: ORDER A FEW DRINKS (first round - 2 of each)
	// Order 2 of each drink type
	// ========================================================================
	fmt.Println("Ordering drinks 👨\u200d🍳⏱️🍻🍻")
	// Order 2 Spritzers (drink_id = 1)
	fmt.Println("\t> Ordering: 2 x Spritzer")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 1, Quantity: 2})
	if err != nil {
		return fmt.Errorf("failed to order Spritzer: %w", err)
	}
	// Order 2 Beers (drink_id = 2)
	fmt.Println("\t> Ordering: 2 x Beer")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 2, Quantity: 2})
	if err != nil {
		return fmt.Errorf("failed to order Beer: %w", err)
	}
	// Order 2 Coffees (drink_id = 3)
	fmt.Println("\t> Ordering: 2 x Coffee")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 3, Quantity: 2})
	if err != nil {
		return fmt.Errorf("failed to order Coffee: %w", err)
	}
	
	// ========================================================================
	// STEP 3: ORDER MORE DRINKS (second round - 6 of each)
	// Order another round of 6 each drink type
	// ========================================================================
	fmt.Println("Ordering another round of drinks 👨\u200d🍳⏱️🍻🍻")
	// Order 6 more Spritzers (drink_id = 1)
	fmt.Println("\t> Ordering: 6 x Spritzer")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 1, Quantity: 6})
	if err != nil {
		return fmt.Errorf("failed to order Spritzer: %w", err)
	}
	// Order 6 more Beers (drink_id = 2)
	fmt.Println("\t> Ordering: 6 x Beer")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 2, Quantity: 6})
	if err != nil {
		return fmt.Errorf("failed to order Beer: %w", err)
	}
	// Order 6 more Coffees (drink_id = 3)
	fmt.Println("\t> Ordering: 6 x Coffee")
	_, err = c.client.OrderDrink(ctx, &pb.OrderRequest{DrinkId: 3, Quantity: 6})
	if err != nil {
		return fmt.Errorf("failed to order Coffee: %w", err)
	}
	
	// ========================================================================
	// STEP 4: GET ORDER TOTAL (THE BILL)
	// Request and display the total of all orders
	// ========================================================================
	fmt.Println("Getting the bill 💹💹💹")
	// Call GetOrders RPC to get all orders
	orders, err := c.client.GetOrders(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get orders: %w", err)
	}
	// Iterate and print each order item with total quantity
	for _, order := range orders.Orders {
		fmt.Printf("\t> Total: %d x %s\n", order.Quantity, order.Drink.Name)
	}
	
	// Return nil to indicate success
	return nil
}
