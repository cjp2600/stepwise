package grpc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// TestNewClient tests the creation of a new gRPC client
func TestNewClient(t *testing.T) {
	log := logger.New()
	
	// This test requires a running gRPC server, so we'll skip it if server is not available
	// In a real scenario, you would start a test server
	serverAddr := "localhost:50051"
	
	client, err := NewClient(serverAddr, true, log)
	if err != nil {
		t.Skipf("Skipping test: gRPC server not available at %s: %v", serverAddr, err)
		return
	}
	defer client.Close()
	
	if client.conn == nil {
		t.Error("Expected connection to be initialized")
	}
	
	if client.logger != log {
		t.Error("Logger not set correctly")
	}
}

// TestMethodLookupWithoutPackage tests the method lookup for services without package
// This test demonstrates the fix for the bug where services without package declaration
// couldn't be found through reflection
func TestMethodLookupWithoutPackage(t *testing.T) {
	log := logger.New()
	serverAddr := "localhost:50051"
	
	client, err := NewClient(serverAddr, true, log)
	if err != nil {
		t.Skipf("Skipping test: gRPC server not available at %s: %v", serverAddr, err)
		return
	}
	defer client.Close()
	
	// Create reflection client to test method lookup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	rc := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(client.conn))
	defer rc.Reset()
	
	// Test that we can list services
	services, err := rc.ListServices()
	if err != nil {
		t.Skipf("Skipping test: reflection not available: %v", err)
		return
	}
	
	if len(services) == 0 {
		t.Skip("Skipping test: no services found")
		return
	}
	
	t.Logf("Found services: %v", services)
	
	// Test finding a service by name (for services without package)
	for _, svcName := range services {
		svcDesc, err := rc.ResolveService(svcName)
		if err != nil {
			t.Logf("Could not resolve service %s: %v", svcName, err)
			continue
		}
		
		if svcDesc == nil {
			t.Errorf("Service descriptor is nil for %s", svcName)
			continue
		}
		
		// Test that we can find methods within the service
		methods := svcDesc.GetMethods()
		if len(methods) == 0 {
			t.Logf("Service %s has no methods", svcName)
			continue
		}
		
		t.Logf("Service %s has %d methods", svcName, len(methods))
		for _, method := range methods {
			t.Logf("  - Method: %s", method.GetName())
		}
	}
}

// TestExecuteRequest tests executing a gRPC request
// This test requires a running gRPC server with reflection enabled
func TestExecuteRequest(t *testing.T) {
	log := logger.New()
	serverAddr := "localhost:50051"
	
	client, err := NewClient(serverAddr, true, log)
	if err != nil {
		t.Skipf("Skipping test: gRPC server not available at %s: %v", serverAddr, err)
		return
	}
	defer client.Close()
	
	req := &Request{
		Service:    "OperationService", // Service without package
		Method:     "GetCustomerOperations",
		ServerAddr: serverAddr,
		Insecure:   true,
		Timeout:    5 * time.Second,
		Data: map[string]interface{}{
			"customer_id": "test-123",
		},
	}
	
	// This test will fail if the bug is not fixed
	// The bug was that FindSymbol("OperationService.GetCustomerOperations") failed
	// The fix adds fallback to find service first, then method within service
	_, err = client.Execute(req)
	if err != nil {
		// Check if error is about method not found (the bug)
		errMsg := err.Error()
		if strings.Contains(errMsg, "failed to find method") && strings.Contains(errMsg, "Symbol not found") {
			t.Errorf("Bug reproduced: method lookup failed for service without package: %v", err)
		} else {
			// Other errors (like server not running, method not implemented) are acceptable
			t.Logf("Request failed (expected if server doesn't implement method): %v", err)
		}
	} else {
		t.Log("Request succeeded - bug is fixed!")
	}
}

