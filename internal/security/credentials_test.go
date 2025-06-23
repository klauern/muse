package security

import (
	"strings"
	"testing"
)

func TestMaskCredential(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty credential",
			input:    "",
			expected: "[not-set]",
		},
		{
			name:     "short credential",
			input:    "abc123",
			expected: "******",
		},
		{
			name:     "normal API key",
			input:    "sk-1234567890abcdef1234567890abcdef",
			expected: "sk-1***************************cdef",
		},
		{
			name:     "long token",
			input:    "ghp_1234567890abcdef1234567890abcdef1234567890",
			expected: "ghp_**************************************7890",
		},
		{
			name:     "exactly 8 chars",
			input:    "12345678",
			expected: "********",
		},
		{
			name:     "exactly 9 chars",
			input:    "123456789",
			expected: "1234*6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskCredential(tt.input)
			if result != tt.expected {
				t.Errorf("MaskCredential(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "api key in config",
			input:    "api_key=sk-1234567890abcdef",
			expected: "api_key=[MASKED]",
		},
		{
			name:     "token in yaml",
			input:    "token: ghp_abcdef123456",
			expected: "token: [MASKED]",
		},
		{
			name:     "password field",
			input:    "password=mysecret123",
			expected: "password=[MASKED]",
		},
		{
			name:     "no sensitive data",
			input:    "model=gpt-4\ntimeout=30s",
			expected: "model=gpt-4\ntimeout=30s",
		},
		{
			name:     "multiple sensitive fields",
			input:    "api_key=secret1\ntoken=secret2\nmodel=gpt-4",
			expected: "api_key=[MASKED]\ntoken=[MASKED]\nmodel=gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLog(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForLog(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateCredential(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorType string
	}{
		{
			name:      "empty credential",
			input:     "",
			wantError: false,
		},
		{
			name:      "valid API key",
			input:     "sk-1234567890abcdef1234567890abcdef",
			wantError: false,
		},
		{
			name:      "placeholder value",
			input:     "your-api-key-here",
			wantError: true,
			errorType: "placeholder",
		},
		{
			name:      "too short",
			input:     "short",
			wantError: true,
			errorType: "too_short",
		},
		{
			name:      "changeme placeholder",
			input:     "changeme",
			wantError: true,
			errorType: "placeholder",
		},
		{
			name:      "valid token",
			input:     "ghp_1234567890abcdef1234567890abcdef",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredential(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateCredential(%q) expected error, got nil", tt.input)
					return
				}
				credErr, ok := err.(*CredentialError)
				if !ok {
					t.Errorf("ValidateCredential(%q) expected CredentialError, got %T", tt.input, err)
					return
				}
				if credErr.Type != tt.errorType {
					t.Errorf("ValidateCredential(%q) error type = %q, want %q", tt.input, credErr.Type, tt.errorType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCredential(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestMaskCredential_DoesNotExposeOriginal(t *testing.T) {
	sensitive := "sk-very-secret-api-key-12345"
	masked := MaskCredential(sensitive)

	// Ensure the masked version doesn't contain the full original
	if strings.Contains(masked, "very-secret-api-key") {
		t.Errorf("Masked credential still contains sensitive portion: %s", masked)
	}

	// Ensure it's actually masked
	if masked == sensitive {
		t.Errorf("Credential was not masked: %s", masked)
	}

	// Ensure it follows expected pattern
	if !strings.HasPrefix(masked, "sk-v") || !strings.HasSuffix(masked, "2345") {
		t.Errorf("Masked credential doesn't follow expected pattern: %s", masked)
	}
}

func TestSanitizeForLog_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"uppercase", "API_KEY=secret"},
		{"lowercase", "api_key=secret"},
		{"mixed case", "Api_Key=secret"},
		{"camelcase", "apiKey=secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLog(tt.input)
			if strings.Contains(result, "secret") {
				t.Errorf("SanitizeForLog(%q) still contains 'secret': %s", tt.input, result)
			}
			if !strings.Contains(result, "[MASKED]") {
				t.Errorf("SanitizeForLog(%q) doesn't contain [MASKED]: %s", tt.input, result)
			}
		})
	}
}
