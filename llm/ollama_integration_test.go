package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klauern/muse/config"
)

func TestOllamaIntegration(t *testing.T) {
	// Mock server to simulate Ollama API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected to request '/api/chat', got: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		var req OllamaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Model != "solar-pro:latest" {
			t.Errorf("Expected model 'solar-pro:latest', got: %s", req.Model)
		}
		if req.Stream {
			t.Error("Expected Stream to be false")
		}
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got: %d", len(req.Messages))
		}
		if len(req.Tools) != 1 {
			t.Errorf("Expected 1 tool, got: %d", len(req.Tools))
		}

		// Simulate Ollama response
		resp := OllamaResponse{
			Message: OllamaMessage{
				Role:    "assistant",
				Content: "feat(test): Add integration test for Ollama",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create Ollama service with mock server URL
	cfg := &config.LLMConfig{
		Provider: "ollama",
		Config: map[string]interface{}{
			"base_url": server.URL,
			"model":    "solar-pro:latest",
		},
	}

	service, err := NewLLMService(cfg)
	if err != nil {
		t.Fatalf("Failed to create LLM service: %v", err)
	}

	// Test GenerateCommitMessage
	ctx := context.Background()
	diff := "diff --git a/test.go b/test.go\n--- a/test.go\n+++ b/test.go\n@@ -1,3 +1,4 @@\n package main\n \n+// This is a test"
	context := "Adding a comment to the test file"
	style := DefaultStyle

	message, err := service.GenerateCommitMessage(ctx, diff, context, style)
	if err != nil {
		t.Fatalf("GenerateCommitMessage failed: %v", err)
	}
	if message != "feat(test): Add integration test for Ollama" {
		t.Errorf("Expected message 'feat(test): Add integration test for Ollama', got: %s", message)
	}
}
