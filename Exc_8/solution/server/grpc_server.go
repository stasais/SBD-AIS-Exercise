package server

import (
	"context"
	"exc8/pb"
	// "fmt" // commented out - was unused import
	"log/slog"
	"net"
	// sync package needed for thread-safe access to orders map
	"sync"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// ============================================================================
// GRPC SERVICE STRUCT
// Implements the OrderServiceServer interface generated from protobuf
// ============================================================================
type GRPCService struct {
	// Embedding UnimplementedOrderServiceServer for forward compatibility
	// This ensures that if new RPC methods are added to the proto, the server still compiles
	pb.UnimplementedOrderServiceServer
	// In-memory storage for available drinks
	// Key is drink ID, value is the Drink object
	drinks map[int32]*pb.Drink
	// In-memory storage for orders
	// Key is drink ID, value is quantity ordered
	orders map[int32]int32
	// Mutex for thread-safe access to orders map
	// Required because multiple goroutines may access orders concurrently
	mu sync.Mutex
}

// ============================================================================
// INITIALIZE DRINKS STORAGE
// Prepopulates the drinks map with available drinks
// ============================================================================
// initDrinks creates and populates the drinks map with predefined drinks
func (s *GRPCService) initDrinks() {
	// Initialize the drinks map
	s.drinks = make(map[int32]*pb.Drink)
	// Initialize the orders map
	s.orders = make(map[int32]int32)
	
	// Add Spritzer drink - id:1, price:2
	s.drinks[1] = &pb.Drink{
		Id:          1,
		Name:        "Spritzer",
		Price:       2,
		Description: "Wine with soda",
	}
	// Add Beer drink - id:2, price:3
	s.drinks[2] = &pb.Drink{
		Id:          2,
		Name:        "Beer",
		Price:       3,
		Description: "Hagenberger Gold",
	}
	// Add Coffee drink - id:3, price:0 (free as shown in expected output)
	s.drinks[3] = &pb.Drink{
		Id:          3,
		Name:        "Coffee",
		Price:       0,
		Description: "Mifare isn't that secure",
	}
}

func StartGrpcServer() error {
	// Create a new gRPC server.
	srv := grpc.NewServer()
	// Create grpc service
	grpcService := &GRPCService{}
	// Initialize drinks storage with prepopulated drinks
	grpcService.initDrinks()
	// Register our service implementation with the gRPC server.
	pb.RegisterOrderServiceServer(srv, grpcService)
	// Serve gRPC server on port 4000.
	lis, err := net.Listen("tcp", ":4000")
	if err != nil {
		return err
	}
	// Log server startup for debugging purposes
	slog.Info("gRPC server started", "port", 4000)
	err = srv.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// todo implement functions

// ============================================================================
// GET DRINKS RPC IMPLEMENTATION
// Returns list of all available drinks
// ============================================================================
// GetDrinks implements the GetDrinks RPC method
// Takes empty request and returns all available drinks
func (s *GRPCService) GetDrinks(ctx context.Context, req *emptypb.Empty) (*pb.DrinkList, error) {
	// Create response object to hold list of drinks
	response := &pb.DrinkList{
		Drinks: make([]*pb.Drink, 0, len(s.drinks)),
	}
	// Iterate over all drinks in the map and add to response
	for _, drink := range s.drinks {
		response.Drinks = append(response.Drinks, drink)
	}
	// Return the list of drinks
	return response, nil
}

// ============================================================================
// ORDER DRINK RPC IMPLEMENTATION
// Processes a drink order request
// ============================================================================
// OrderDrink implements the OrderDrink RPC method
// Takes order request with drink_id and quantity, returns boolean success
func (s *GRPCService) OrderDrink(ctx context.Context, req *pb.OrderRequest) (*wrapperspb.BoolValue, error) {
	// Check if the requested drink exists
	drink, exists := s.drinks[req.DrinkId]
	if !exists {
		// Drink not found, return false
		slog.Warn("Drink not found", "drink_id", req.DrinkId)
		return wrapperspb.Bool(false), nil
	}
	
	// Lock mutex for thread-safe access to orders map
	s.mu.Lock()
	// Defer unlock to ensure mutex is released after function returns
	defer s.mu.Unlock()
	
	// Add the ordered quantity to existing orders for this drink
	s.orders[req.DrinkId] += req.Quantity
	
	// Log the order for debugging
	slog.Info("Order placed", "drink", drink.Name, "quantity", req.Quantity)
	
	// Return success
	return wrapperspb.Bool(true), nil
}

// ============================================================================
// GET ORDERS RPC IMPLEMENTATION
// Returns all orders (the bill)
// ============================================================================
// GetOrders implements the GetOrders RPC method
// Takes empty request and returns all current orders
func (s *GRPCService) GetOrders(ctx context.Context, req *emptypb.Empty) (*pb.OrderList, error) {
	// Lock mutex for thread-safe access to orders map
	s.mu.Lock()
	// Defer unlock to ensure mutex is released after function returns
	defer s.mu.Unlock()
	
	// Create response object to hold list of orders
	response := &pb.OrderList{
		Orders: make([]*pb.OrderItem, 0),
	}
	
	// Iterate over all orders and create OrderItem for each
	for drinkId, quantity := range s.orders {
		// Skip if no orders for this drink
		if quantity == 0 {
			continue
		}
		// Get the drink details
		drink := s.drinks[drinkId]
		// Create order item with drink and quantity
		orderItem := &pb.OrderItem{
			Drink:    drink,
			Quantity: quantity,
		}
		// Add to response
		response.Orders = append(response.Orders, orderItem)
	}
	
	// Return the list of orders
	return response, nil
}
