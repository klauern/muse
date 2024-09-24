package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klauern/muse/config"
	"github.com/stretchr/testify/assert"
)

func TestOllamaIntegration(t *testing.T) {
	// Mock server to simulate Ollama API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/chat", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req OllamaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		assert.Equal(t, "solar-pro:latest", req.Model)
		assert.False(t, req.Stream)
		assert.Len(t, req.Messages, 2)
		assert.Len(t, req.Tools, 1)

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
		Ollama: config.OllamaConfig{
			BaseURL: server.URL,
			Model:   "solar-pro:latest",
		},
	}

	service, err := NewLLMService(cfg)
	assert.NoError(t, err)

	// Test GenerateCommitMessage
	ctx := context.Background()
	diff := "diff --git a/test.go b/test.go\n--- a/test.go\n+++ b/test.go\n@@ -1,3 +1,4 @@\n package main\n \n+// This is a test"
	context := "Adding a comment to the test file"
	style := DefaultStyle

	message, err := service.GenerateCommitMessage(ctx, diff, context, style)
	assert.NoError(t, err)
	assert.Equal(t, "feat(test): Add integration test for Ollama", message)
}
