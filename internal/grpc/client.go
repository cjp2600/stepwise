package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client represents a gRPC client for making requests
type Client struct {
	conn   *grpc.ClientConn
	logger *logger.Logger
}

// Request represents a gRPC request
type Request struct {
	Service    string            `yaml:"service" json:"service"`
	Method     string            `yaml:"method" json:"method"`
	Data       interface{}       `yaml:"data" json:"data"`
	Metadata   map[string]string `yaml:"metadata" json:"metadata"`
	Timeout    time.Duration     `yaml:"timeout" json:"timeout"`
	Insecure   bool              `yaml:"insecure" json:"insecure"`
	ServerAddr string            `yaml:"server_addr" json:"server_addr"`
}

// Response represents a gRPC response
type Response struct {
	Data       interface{}         `json:"data"`
	Metadata   map[string][]string `json:"metadata"`
	Duration   time.Duration       `json:"duration"`
	Error      error               `json:"error,omitempty"`
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
}

// NewClient creates a new gRPC client
func NewClient(serverAddr string, useInsecure bool, log *logger.Logger) (*Client, error) {
	var opts []grpc.DialOption

	if useInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &Client{
		conn:   conn,
		logger: log,
	}, nil
}

// Execute performs a gRPC request
func (c *Client) Execute(req *Request) (*Response, error) {
	start := time.Now()

	c.logger.Debug("Making gRPC request",
		"service", req.Service,
		"method", req.Method,
		"server", req.ServerAddr)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()

	// Add metadata if provided
	if len(req.Metadata) > 0 {
		md := metadata.New(nil)
		for key, value := range req.Metadata {
			md.Set(key, value)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Mock gRPC service responses for testing
	var responseData interface{}

	switch req.Service {
	case "UserService":
		switch req.Method {
		case "GetUser":
			// Mock user service response
			responseData = map[string]interface{}{
				"user_id": req.Data.(map[string]interface{})["user_id"],
				"name":    "John Doe",
				"email":   "john.doe@example.com",
				"status":  "active",
			}
		default:
			return nil, fmt.Errorf("unknown method %s for service %s", req.Method, req.Service)
		}
	case "OrderService":
		switch req.Method {
		case "CreateOrder":
			// Mock order service response
			responseData = map[string]interface{}{
				"order_id":     "ORD-12345",
				"user_id":      req.Data.(map[string]interface{})["user_id"],
				"status":       "created",
				"total_amount": req.Data.(map[string]interface{})["total_amount"],
				"items":        req.Data.(map[string]interface{})["items"],
			}
		default:
			return nil, fmt.Errorf("unknown method %s for service %s", req.Method, req.Service)
		}
	default:
		return nil, fmt.Errorf("unknown service %s", req.Service)
	}

	duration := time.Since(start)

	response := &Response{
		Data:       responseData,
		Metadata:   make(map[string][]string),
		Duration:   duration,
		Status:     "OK",
		StatusCode: 0, // gRPC OK status
	}

	c.logger.Debug("Received gRPC response",
		"service", req.Service,
		"method", req.Method,
		"duration", duration,
		"status", response.Status)

	return response, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsSuccess returns true if the response indicates success
func (r *Response) IsSuccess() bool {
	return r.Error == nil && r.StatusCode == 0
}

// GetData returns the response data
func (r *Response) GetData() interface{} {
	return r.Data
}

// GetMetadata returns the response metadata
func (r *Response) GetMetadata() map[string][]string {
	return r.Metadata
}

// GetDuration returns the response duration
func (r *Response) GetDuration() time.Duration {
	return r.Duration
}

// GetError returns the response error
func (r *Response) GetError() error {
	return r.Error
}
