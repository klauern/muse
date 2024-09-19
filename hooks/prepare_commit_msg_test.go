package hooks

import (
	"context"
	"os"
	"testing"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// LLMHook is defined in another file, so we'll just declare a mock version for testing
type MockLLMHook struct {
	Generator llm.Generator
	Config    *config.Config
}

func (h *MockLLMHook) Run(commitMsgFile, commitSource, commitSHA string) error {
	// Mock implementation for testing
	return nil
}

// MockCommitMessageGenerator is a mock for the CommitMessageGenerator
type MockCommitMessageGenerator struct {
	mock.Mock
}

func (m *MockCommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle string) (string, error) {
	args := m.Called(ctx, diff, commitStyle)
	return args.String(0), args.Error(1)
}

// Ensure MockCommitMessageGenerator implements llm.Generator
var _ llm.Generator = (*MockCommitMessageGenerator)(nil)

func TestLLMHook_Run(t *testing.T) {
	// Create a temporary file for the commit message
	tmpfile, err := os.CreateTemp("", "commit-msg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a mock generator
	mockGenerator := new(MockCommitMessageGenerator)

	// Create the LLMHook
	hook := &LLMHook{
		Generator: mockGenerator,
		Config: &config.Config{
			Hook: config.Hook{
				CommitStyle: "conventional",
				DryRun:      false,
				Preview:     false,
			},
		},
	}

	// Set up the mock expectation
	mockGenerator.On("Generate", mock.Anything, mock.Anything, "conventional").Return("feat: test commit message", nil)

	// Run the hook
	err = hook.Run(tmpfile.Name(), "", "")

	// Assert that there was no error
	assert.NoError(t, err)

	// Read the content of the commit message file
	content, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Assert that the content matches the generated message
	assert.Equal(t, "feat: test commit message", string(content))

	// Assert that the mock method was called
	mockGenerator.AssertExpectations(t)
}

func TestLLMHook_Run_DryRun(t *testing.T) {
	mockGenerator := new(MockCommitMessageGenerator)
	hook := &LLMHook{
		Generator: mockGenerator,
		Config: &config.Config{
			Hook: config.Hook{
				CommitStyle: "conventional",
				DryRun:      true,
				Preview:     false,
			},
		},
	}

	mockGenerator.On("Generate", mock.Anything, mock.Anything, "conventional").Return("feat: test commit message", nil)

	err := hook.Run("dummy-file", "", "")

	assert.NoError(t, err)
	mockGenerator.AssertExpectations(t)
}

func TestDefaultHook_Run(t *testing.T) {
	hook := &DefaultHook{}

	err := hook.Run("dummy-file", "", "")

	assert.NoError(t, err)
}
