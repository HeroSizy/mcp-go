package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMCPServer_NewMCPServer(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0")
	assert.NotNil(t, server)
	assert.Equal(t, "test-server", server.name)
	assert.Equal(t, "1.0.0", server.version)
}

func TestMCPServer_Capabilities(t *testing.T) {
	tests := []struct {
		name     string
		options  []ServerOption
		validate func(t *testing.T, response mcp.JSONRPCMessage)
	}{
		{
			name:    "No capabilities",
			options: []ServerOption{},
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				initResult, ok := resp.Result.(mcp.InitializeResult)
				assert.True(t, ok)

				assert.Equal(
					t,
					mcp.LATEST_PROTOCOL_VERSION,
					initResult.ProtocolVersion,
				)
				assert.Equal(t, "test-server", initResult.ServerInfo.Name)
				assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)
				assert.Nil(t, initResult.Capabilities.Resources)
				assert.Nil(t, initResult.Capabilities.Prompts)
				assert.Nil(t, initResult.Capabilities.Tools)
				assert.Nil(t, initResult.Capabilities.Logging)
			},
		},
		{
			name: "All capabilities",
			options: []ServerOption{
				WithResourceCapabilities(true, true),
				WithPromptCapabilities(true),
				WithToolCapabilities(true),
				WithLogging(),
			},
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				initResult, ok := resp.Result.(mcp.InitializeResult)
				assert.True(t, ok)

				assert.Equal(
					t,
					mcp.LATEST_PROTOCOL_VERSION,
					initResult.ProtocolVersion,
				)
				assert.Equal(t, "test-server", initResult.ServerInfo.Name)
				assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)

				assert.NotNil(t, initResult.Capabilities.Resources)
				if initResult.Capabilities.Resources != nil {
					assert.True(t, initResult.Capabilities.Resources.Subscribe)
					assert.True(
						t,
						initResult.Capabilities.Resources.ListChanged,
					)
				}

				assert.NotNil(t, initResult.Capabilities.Prompts)
				if initResult.Capabilities.Prompts != nil {
					assert.True(t, initResult.Capabilities.Prompts.ListChanged)
				}

				assert.NotNil(t, initResult.Capabilities.Tools)
				if initResult.Capabilities.Tools != nil {
					assert.True(t, initResult.Capabilities.Tools.ListChanged)
				}

				assert.NotNil(t, initResult.Capabilities.Logging)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewMCPServer("test-server", "1.0.0", tt.options...)
			message := mcp.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Request: mcp.Request{
					Method: "initialize",
				},
			}
			messageBytes, err := json.Marshal(message)
			assert.NoError(t, err)

			response := server.HandleMessage(context.Background(), messageBytes)
			tt.validate(t, response)
		})
	}
}

func TestMCPServer_HandleValidMessages(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0",
		WithResourceCapabilities(true, true),
		WithPromptCapabilities(true),
		WithToolCapabilities(true),
	)

	tests := []struct {
		name     string
		message  interface{}
		validate func(t *testing.T, response mcp.JSONRPCMessage)
	}{
		{
			name: "Initialize request",
			message: mcp.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Request: mcp.Request{
					Method: "initialize",
				},
			},
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				initResult, ok := resp.Result.(mcp.InitializeResult)
				assert.True(t, ok)

				assert.Equal(
					t,
					mcp.LATEST_PROTOCOL_VERSION,
					initResult.ProtocolVersion,
				)
				assert.Equal(t, "test-server", initResult.ServerInfo.Name)
				assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)
			},
		},
		{
			name: "Ping request",
			message: mcp.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Request: mcp.Request{
					Method: "ping",
				},
			},
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				_, ok = resp.Result.(mcp.EmptyResult)
				assert.True(t, ok)
			},
		},
		{
			name: "List resources",
			message: mcp.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Request: mcp.Request{
					Method: "resources/list",
				},
			},
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				listResult, ok := resp.Result.(mcp.ListResourcesResult)
				assert.True(t, ok)
				assert.NotNil(t, listResult.Resources)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageBytes, err := json.Marshal(tt.message)
			assert.NoError(t, err)

			response := server.HandleMessage(context.Background(), messageBytes)
			assert.NotNil(t, response)
			tt.validate(t, response)
		})
	}
}

func TestMCPServer_HandlePagination(t *testing.T) {
	server := createTestServer()

	tests := []struct {
		name     string
		message  string
		validate func(t *testing.T, response mcp.JSONRPCMessage)
	}{
		{
			name: "List resources with cursor",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "resources/list",
                    "params": {
                        "cursor": "test-cursor"
                    }
                }`,
			validate: func(t *testing.T, response mcp.JSONRPCMessage) {
				resp, ok := response.(mcp.JSONRPCResponse)
				assert.True(t, ok)

				listResult, ok := resp.Result.(mcp.ListResourcesResult)
				assert.True(t, ok)
				assert.NotNil(t, listResult.Resources)
				assert.Equal(t, mcp.Cursor(""), listResult.NextCursor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.HandleMessage(
				context.Background(),
				[]byte(tt.message),
			)
			tt.validate(t, response)
		})
	}
}

func TestMCPServer_HandleNotifications(t *testing.T) {
	server := createTestServer()
	notificationReceived := false

	server.AddNotificationHandler(func(notification mcp.JSONRPCNotification) {
		notificationReceived = true
	})

	message := `{
            "jsonrpc": "2.0",
            "method": "notifications/initialized"
        }`

	response := server.HandleMessage(context.Background(), []byte(message))
	assert.Nil(t, response)
	assert.True(t, notificationReceived)
}

func TestMCPServer_HandleInvalidMessages(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0")

	tests := []struct {
		name        string
		message     string
		expectedErr int
	}{
		{
			name:        "Invalid JSON",
			message:     `{"jsonrpc": "2.0", "id": 1, "method": "initialize"`,
			expectedErr: mcp.PARSE_ERROR,
		},
		{
			name:        "Invalid method",
			message:     `{"jsonrpc": "2.0", "id": 1, "method": "nonexistent"}`,
			expectedErr: mcp.METHOD_NOT_FOUND,
		},
		{
			name:        "Invalid parameters",
			message:     `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": "invalid"}`,
			expectedErr: mcp.INVALID_REQUEST,
		},
		{
			name:        "Missing JSONRPC version",
			message:     `{"id": 1, "method": "initialize"}`,
			expectedErr: mcp.INVALID_REQUEST,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.HandleMessage(
				context.Background(),
				[]byte(tt.message),
			)
			assert.NotNil(t, response)

			errorResponse, ok := response.(mcp.JSONRPCError)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedErr, errorResponse.Error.Code)
		})
	}
}

func TestMCPServer_HandleUndefinedHandlers(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0",
		WithResourceCapabilities(true, true),
		WithPromptCapabilities(true),
		WithToolCapabilities(true),
	)

	tests := []struct {
		name        string
		message     string
		expectedErr int
	}{
		{
			name: "Undefined tool",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "tools/call",
                    "params": {
                        "name": "undefined-tool",
                        "arguments": {}
                    }
                }`,
			expectedErr: mcp.INVALID_PARAMS,
		},
		{
			name: "Undefined prompt",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "prompts/get",
                    "params": {
                        "name": "undefined-prompt",
                        "arguments": {}
                    }
                }`,
			expectedErr: mcp.INVALID_PARAMS,
		},
		{
			name: "Undefined resource",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "resources/read",
                    "params": {
                        "uri": "undefined-resource"
                    }
                }`,
			expectedErr: mcp.INVALID_PARAMS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.HandleMessage(
				context.Background(),
				[]byte(tt.message),
			)
			assert.NotNil(t, response)

			errorResponse, ok := response.(mcp.JSONRPCError)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedErr, errorResponse.Error.Code)
		})
	}
}

func TestMCPServer_HandleMethodsWithoutCapabilities(t *testing.T) {
	server := NewMCPServer(
		"test-server",
		"1.0.0",
	) // No capabilities enabled

	tests := []struct {
		name        string
		message     string
		expectedErr int
	}{
		{
			name: "Tools without capabilities",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "tools/call",
                    "params": {
                        "name": "test-tool"
                    }
                }`,
			expectedErr: mcp.METHOD_NOT_FOUND,
		},
		{
			name: "Prompts without capabilities",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "prompts/get",
                    "params": {
                        "name": "test-prompt"
                    }
                }`,
			expectedErr: mcp.METHOD_NOT_FOUND,
		},
		{
			name: "Resources without capabilities",
			message: `{
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "resources/read",
                    "params": {
                        "uri": "test-resource"
                    }
                }`,
			expectedErr: mcp.METHOD_NOT_FOUND,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.HandleMessage(
				context.Background(),
				[]byte(tt.message),
			)
			assert.NotNil(t, response)

			errorResponse, ok := response.(mcp.JSONRPCError)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedErr, errorResponse.Error.Code)
		})
	}
}

func createTestServer() *MCPServer {
	server := NewMCPServer("test-server", "1.0.0",
		WithResourceCapabilities(true, true),
		WithPromptCapabilities(true),
		WithToolCapabilities(true),
	)

	server.AddResource(
		"resource://testresource",
		func() ([]interface{}, error) {
			return []interface{}{
				mcp.TextResourceContents{
					ResourceContents: mcp.ResourceContents{
						URI:      "resource://testresource",
						MIMEType: "text/plain",
					},
					Text: "test content",
				},
			}, nil
		},
	)

	server.AddTool(
		mcp.Tool{
			Name:        "test-tool",
			Description: "Test tool",
		},
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []interface{}{
					mcp.TextContent{
						Type: "text",
						Text: "test result",
					},
				},
			}, nil
		},
	)

	return server
}
