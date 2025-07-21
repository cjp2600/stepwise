package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Mock gRPC services for testing
type UserService struct {
	UnimplementedUserServiceServer
}

type OrderService struct {
	UnimplementedOrderServiceServer
}

// Mock service implementations
func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	// Log metadata if present
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("Received metadata: %v", md)
	}

	return &GetUserResponse{
		UserId: req.UserId,
		Name:   "John Doe",
		Email:  "john.doe@example.com",
		Status: "active",
	}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Log metadata if present
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("Received metadata: %v", md)
	}

	return &CreateOrderResponse{
		OrderId:     "ORD-12345",
		UserId:      req.UserId,
		Status:      "created",
		TotalAmount: req.TotalAmount,
		Items:       req.Items,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Register services
	RegisterUserServiceServer(s, &UserService{})
	RegisterOrderServiceServer(s, &OrderService{})

	log.Printf("gRPC test server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Mock protobuf definitions (in a real implementation, these would be generated from .proto files)
type GetUserRequest struct {
	UserId string `json:"user_id"`
}

type GetUserResponse struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type CreateOrderRequest struct {
	UserId      string      `json:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
}

type CreateOrderResponse struct {
	OrderId     string      `json:"order_id"`
	UserId      string      `json:"user_id"`
	Status      string      `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductId string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Mock service interfaces (in a real implementation, these would be generated from .proto files)
type UserServiceServer interface {
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
}

type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
}

type UnimplementedUserServiceServer struct{}

func (UnimplementedUserServiceServer) GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error) {
	return nil, fmt.Errorf("method GetUser not implemented")
}

type UnimplementedOrderServiceServer struct{}

func (UnimplementedOrderServiceServer) CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error) {
	return nil, fmt.Errorf("method CreateOrder not implemented")
}

func RegisterUserServiceServer(s *grpc.Server, srv UserServiceServer) {
	// Mock registration - in real implementation this would be generated
}

func RegisterOrderServiceServer(s *grpc.Server, srv OrderServiceServer) {
	// Mock registration - in real implementation this would be generated
}
