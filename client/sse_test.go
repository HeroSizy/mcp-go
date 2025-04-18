package client

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestSSEMCPClient(t *testing.T) {
	// Create MCP server with capabilities
	mcpServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)

	// Create test server
	testServer := server.NewTestServer(mcpServer)
	defer testServer.Close()

	t.Run("Can create client", func(t *testing.T) {
		client, err := NewSSEMCPClient(testServer.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		if client.baseURL == nil {
			t.Error("Base URL should not be nil")
		}
	})

	t.Run("Can initialize and make requests", func(t *testing.T) {
		client, err := NewSSEMCPClient(testServer.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start the client
		if err := client.Start(ctx); err != nil {
			t.Fatalf("Failed to start client: %v", err)
		}

		// Initialize
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		}

		result, err := client.Initialize(ctx, initRequest)
		if err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}

		if result.ServerInfo.Name != "test-server" {
			t.Errorf(
				"Expected server name 'test-server', got '%s'",
				result.ServerInfo.Name,
			)
		}

		// Test Ping
		if err := client.Ping(ctx); err != nil {
			t.Errorf("Ping failed: %v", err)
		}

		// Test ListTools
		toolsRequest := mcp.ListToolsRequest{}
		_, err = client.ListTools(ctx, toolsRequest)
		if err != nil {
			t.Errorf("ListTools failed: %v", err)
		}
	})

	// t.Run("Can handle notifications", func(t *testing.T) {
	// 	client, err := NewSSEMCPClient(testServer.URL + "/sse")
	// 	if err != nil {
	// 		t.Fatalf("Failed to create client: %v", err)
	// 	}
	// 	defer client.Close()

	// 	notificationReceived := make(chan mcp.JSONRPCNotification, 1)
	// 	client.OnNotification(func(notification mcp.JSONRPCNotification) {
	// 		notificationReceived <- notification
	// 	})

	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()

	// 	if err := client.Start(ctx); err != nil {
	// 		t.Fatalf("Failed to start client: %v", err)
	// 	}

	// 	// Initialize first
	// 	initRequest := mcp.InitializeRequest{}
	// 	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	// 	initRequest.Params.ClientInfo = mcp.Implementation{
	// 		Name:    "test-client",
	// 		Version: "1.0.0",
	// 	}

	// 	_, err = client.Initialize(ctx, initRequest)
	// 	if err != nil {
	// 		t.Fatalf("Failed to initialize: %v", err)
	// 	}

	// 	// Subscribe to a resource to test notifications
	// 	subRequest := mcp.SubscribeRequest{}
	// 	subRequest.Params.URI = "test://resource"
	// 	if err := client.Subscribe(ctx, subRequest); err != nil {
	// 		t.Fatalf("Failed to subscribe: %v", err)
	// 	}

	// 	select {
	// 	case <-notificationReceived:
	// 		// Success
	// 	case <-time.After(time.Second):
	// 		t.Error("Timeout waiting for notification")
	// 	}
	// })

	t.Run("Handles errors properly", func(t *testing.T) {
		client, err := NewSSEMCPClient(testServer.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Start(ctx); err != nil {
			t.Fatalf("Failed to start client: %v", err)
		}

		// Try to make a request without initializing
		toolsRequest := mcp.ListToolsRequest{}
		_, err = client.ListTools(ctx, toolsRequest)
		if err == nil {
			t.Error("Expected error when making request before initialization")
		}
	})

	// t.Run("Handles context cancellation", func(t *testing.T) {
	// 	client, err := NewSSEMCPClient(testServer.URL + "/sse")
	// 	if err != nil {
	// 		t.Fatalf("Failed to create client: %v", err)
	// 	}
	// 	defer client.Close()

	// 	if err := client.Start(context.Background()); err != nil {
	// 		t.Fatalf("Failed to start client: %v", err)
	// 	}

	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	cancel() // Cancel immediately

	// 	toolsRequest := mcp.ListToolsRequest{}
	// 	_, err = client.ListTools(ctx, toolsRequest)
	// 	if err == nil {
	// 		t.Error("Expected error when context is cancelled")
	// 	}
	// })
}
