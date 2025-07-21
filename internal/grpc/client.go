package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"bytes"
	"encoding/json"

	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/jhump/protoreflect/dynamic"
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
	var headers []string
	if len(req.Metadata) > 0 {
		for key, value := range req.Metadata {
			headers = append(headers, key+":"+value)
		}
	}

	// Reflection client
	rc := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(c.conn))
	descSource := grpcurl.DescriptorSourceFromServer(ctx, rc)
	defer rc.Reset()

	methodName := req.Service + "." + req.Method
	c.logger.Debug("Looking up method", "method", methodName)

	mDesc, err := descSource.FindSymbol(methodName)
	if err != nil {
		c.logger.Debug("Method lookup failed", "error", err)
		return nil, fmt.Errorf("failed to find method %s: %w", methodName, err)
	}
	method, ok := mDesc.(*desc.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a method", methodName)
	}

	inputType := method.GetInputType()
	c.logger.Debug("Input type", "fullName", inputType.GetFullyQualifiedName())

	// Marshal req.Data to JSON и затем в dynamic.Message
	jsonData, err := json.Marshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	msg := dynamic.NewMessage(inputType)
	if err := msg.UnmarshalJSON(jsonData); err != nil {
		c.logger.Debug("UnmarshalJSON error", "error", err)
		return nil, fmt.Errorf("failed to unmarshal data to proto: %w", err)
	}

	// Prepare metadata
	md := make(map[string]string)
	for _, h := range headers {
		parts := bytes.SplitN([]byte(h), []byte{':'}, 2)
		if len(parts) == 2 {
			md[string(parts[0])] = string(parts[1])
		}
	}

	// Prepare output message
	outputType := method.GetOutputType()
	outMsg := dynamic.NewMessage(outputType)

	c.logger.Debug("Invoking gRPC method", "method", methodName)
	fullMethod := fmt.Sprintf("/%s/%s", req.Service, req.Method)
	c.logger.Debug("Full gRPC method path", "fullMethod", fullMethod)
	err = c.conn.Invoke(ctx, fullMethod, msg, outMsg)
	if err != nil {
		c.logger.Debug("Invoke error", "error", err)
		return nil, fmt.Errorf("gRPC invoke error: %w", err)
	}

	// Marshal response to JSON
	jsonResp, err := outMsg.MarshalJSON()
	if err != nil {
		c.logger.Debug("MarshalJSON error", "error", err)
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var respData map[string]interface{}
	if err := json.Unmarshal(jsonResp, &respData); err != nil {
		c.logger.Debug("Unmarshal error", "error", err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	duration := time.Since(start)
	response := &Response{
		Data:       respData,
		Metadata:   make(map[string][]string),
		Duration:   duration,
		Status:     "OK",
		StatusCode: 0,
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
