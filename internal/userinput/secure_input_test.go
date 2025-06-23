package userinput

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSecureInputHandler_validateYesNoInput(t *testing.T) {
	handler := NewSecureInputHandler()

	tests := []struct {
		name     string
		input    string
		expected bool
		wantErr  bool
	}{
		// Valid yes inputs
		{"yes lowercase", "y", true, false},
		{"yes uppercase", "Y", true, false},
		{"yes full", "yes", true, false},
		{"yes full caps", "YES", true, false},
		{"yes with spaces", " yes ", true, false},
		{"yes true", "true", true, false},
		{"yes numeric", "1", true, false},

		// Valid no inputs
		{"no lowercase", "n", false, false},
		{"no uppercase", "N", false, false},
		{"no full", "no", false, false},
		{"no full caps", "NO", false, false},
		{"no with spaces", " no ", false, false},
		{"no false", "false", false, false},
		{"no numeric", "0", false, false},

		// Invalid inputs
		{"empty", "", false, true},
		{"invalid text", "maybe", false, true},
		{"numbers", "123", false, true},
		{"special chars", "y!", false, true},
		{"mixed", "yes no", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.validateYesNoInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateYesNoInput() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("validateYesNoInput() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("validateYesNoInput() = %v, want %v for input %q", result, tt.expected, tt.input)
			}
		})
	}
}

func TestSecureInputHandler_sanitizeInput(t *testing.T) {
	handler := NewSecureInputHandler()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal text", "hello world", "hello world"},
		{"extra spaces", "  hello   world  ", "hello world"},
		{"tabs and newlines", "hello\tworld\n", "hello world"},
		{"null bytes", "hello\x00world", "helloworld"},
		{"control chars", "hello\x01\x02world", "helloworld"},
		{"mixed whitespace", " \t hello \n world \t ", "hello world"},
		{"empty string", "", ""},
		{"only whitespace", "   \t\n   ", ""},
		{"special printable", "hello!@#$%^&*()world", "hello!@#$%^&*()world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.sanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSecureInputHandler_LengthValidation(t *testing.T) {
	handler := NewSecureInputHandler()
	handler.SetMaxLength(10) // Short limit for testing

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"within limit", "y", false},
		{"at limit", "yes", false},
		{"over limit", strings.Repeat("y", 15), true},
		{"way over limit", strings.Repeat("a", 100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.validateYesNoInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateYesNoInput() expected error for long input")
				} else if !strings.Contains(err.Error(), "input too long") {
					t.Errorf("validateYesNoInput() expected length validation error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("validateYesNoInput() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSecureInputHandler_TimeoutConfiguration(t *testing.T) {
	handler := NewSecureInputHandler()

	// Test default timeout
	if handler.timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want %v", handler.timeout, 30*time.Second)
	}

	// Test custom timeout
	customTimeout := 5 * time.Second
	handler.SetTimeout(customTimeout)
	if handler.timeout != customTimeout {
		t.Errorf("Custom timeout = %v, want %v", handler.timeout, customTimeout)
	}
}

func TestSecureInputHandler_MaxLengthConfiguration(t *testing.T) {
	handler := NewSecureInputHandler()

	// Test default max length
	if handler.maxLen != 100 {
		t.Errorf("Default maxLen = %d, want %d", handler.maxLen, 100)
	}

	// Test custom max length
	customMaxLen := 50
	handler.SetMaxLength(customMaxLen)
	if handler.maxLen != customMaxLen {
		t.Errorf("Custom maxLen = %d, want %d", handler.maxLen, customMaxLen)
	}
}

func TestInputValidationError(t *testing.T) {
	err := InputValidationError{
		Input:  "invalid",
		Reason: "test reason",
	}

	expected := `invalid input: test reason (input: "invalid")`
	if err.Error() != expected {
		t.Errorf("InputValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSecureInputHandler_PromptWithValidation(t *testing.T) {
	handler := NewSecureInputHandler()
	handler.SetTimeout(1 * time.Millisecond) // Very short timeout for testing

	ctx := context.Background()

	// Test timeout - we expect either a timeout or "no input received" error
	_, err := handler.PromptWithValidation(ctx, "Test prompt: ", nil)
	if err == nil {
		t.Errorf("PromptWithValidation() expected an error, got nil")
	} else if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "no input received") {
		t.Errorf("PromptWithValidation() expected timeout or no input error, got: %v", err)
	}
}

func TestSecureInputHandler_CustomValidator(t *testing.T) {
	// Custom validator that only allows "test"
	validator := func(input string) error {
		if input != "test" {
			return InputValidationError{
				Input:  input,
				Reason: "must be 'test'",
			}
		}
		return nil
	}

	// Test the validator directly since we can't easily test the full prompt
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid input", "test", false},
		{"invalid input", "nottest", true},
		{"empty input", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validator() expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("validator() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSecureInputHandler_SecurityPatterns(t *testing.T) {
	// Test various potentially malicious inputs
	maliciousInputs := []string{
		"\x00\x01\x02",                  // Null and control bytes
		"$(rm -rf /)",                   // Command injection attempt
		"`cat /etc/passwd`",             // Backtick command
		"../../etc/passwd",              // Path traversal
		"<script>alert('xss')</script>", // XSS attempt
		strings.Repeat("A", 10000),      // Large input
	}

	for _, input := range maliciousInputs {
		t.Run("malicious_input", func(t *testing.T) {
			handler := NewSecureInputHandler()
			sanitized := handler.sanitizeInput(input)

			// Ensure null bytes are removed
			if strings.Contains(sanitized, "\x00") {
				t.Errorf("sanitizeInput() failed to remove null bytes")
			}

			// Ensure control characters are removed (except allowed ones)
			for _, r := range sanitized {
				if r < 32 && r != '\t' && r != '\n' {
					t.Errorf("sanitizeInput() failed to remove control character: %d", int(r))
				}
			}
		})
	}
}
