package llm

import (
	"context"
	"os"
	"testing"

	"github.com/klauern/muse/templates"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIServiceIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	service, err := NewOpenAIService(apiKey, "gpt-3.5-turbo")
	assert.NoError(t, err)

	diff := "diff --git a/file.txt b/file.txt\nindex 83db48f..bf269f4 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-Hello World\n+Hello OpenAI"
	style := templates.Conventional

	ctx := context.Background()
	message, err := service.GenerateCommitMessage(ctx, diff, style)
	assert.NoError(t, err)
	assert.NotEmpty(t, message)
}
