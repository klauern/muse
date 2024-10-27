package llm

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/klauern/muse/templates"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCommitMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockOpenAIServiceInterface(ctrl)
	generator := &CommitMessageGenerator{LLMService: mockService}

	diff := "diff --git a/file.txt b/file.txt\nindex 83db48f..bf269f4 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-Hello World\n+Hello OpenAI"
	style := templates.Conventional

	mockService.EXPECT().GenerateCommitMessage(gomock.Any(), diff, style).Return("feat: update greeting message", nil)

	ctx := context.Background()
	message, err := generator.Generate(ctx, diff, style)
	assert.NoError(t, err)
	assert.Equal(t, "feat: update greeting message", message)
}
