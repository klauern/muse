package security

import (
	"strings"
)

// MaskCredential masks sensitive credentials for safe logging
// Shows first 4 and last 4 characters with asterisks in between
func MaskCredential(credential string) string {
	if credential == "" {
		return "[not-set]"
	}

	// For very short credentials, mask completely
	if len(credential) <= 8 {
		return strings.Repeat("*", len(credential))
	}

	// For 9+ characters, show first 4, middle asterisks, last 4
	// But ensure we don't have overlapping regions
	prefixLen := 4
	suffixLen := 4

	// If the credential is very short, reduce the suffix length
	if len(credential) == 9 {
		suffixLen = 4 // "123456789" -> "1234*5789"
		// We want indices: 0-3 (prefix), 4 (asterisk), 5-8 (suffix)
		prefix := credential[:prefixLen]
		suffix := credential[prefixLen+1:] // Skip the middle character
		return prefix + "*" + suffix
	}

	// For longer credentials, use standard approach
	prefix := credential[:prefixLen]
	suffix := credential[len(credential)-suffixLen:]
	middleLength := len(credential) - prefixLen - suffixLen
	middle := strings.Repeat("*", middleLength)

	return prefix + middle + suffix
}

// SanitizeForLog removes sensitive information from strings for logging
func SanitizeForLog(input string) string {
	// Common patterns for API keys, tokens, passwords
	sensitivePatterns := []string{
		"api_key", "apikey", "api-key",
		"token", "password", "passwd", "pwd",
		"secret", "auth", "bearer",
	}

	result := input
	for _, pattern := range sensitivePatterns {
		// Case-insensitive replacement
		lowerInput := strings.ToLower(result)
		lowerPattern := strings.ToLower(pattern)

		if strings.Contains(lowerInput, lowerPattern) {
			// Find and mask the value after the pattern
			result = maskSensitiveValues(result, pattern)
		}
	}

	return result
}

// maskSensitiveValues looks for key=value or key:value patterns and masks the values
func maskSensitiveValues(input, sensitiveKey string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		lowerKey := strings.ToLower(sensitiveKey)

		if strings.Contains(lowerLine, lowerKey) {
			// Look for patterns like key=value or key: value
			if idx := strings.Index(lowerLine, "="); idx > 0 {
				keyPart := line[:idx+1]
				lines[i] = keyPart + "[MASKED]"
			} else if idx := strings.Index(lowerLine, ":"); idx > 0 {
				keyPart := line[:idx+1]
				lines[i] = keyPart + " [MASKED]"
			}
		}
	}
	return strings.Join(lines, "\n")
}

// ValidateCredential performs basic validation on credentials
func ValidateCredential(credential string) error {
	if credential == "" {
		return nil // Empty is handled by callers
	}

	// Check for obvious placeholder values
	placeholders := []string{
		"your-api-key", "your-token", "your-secret",
		"api-key-here", "token-here", "secret-here",
		"changeme", "replace-me", "todo", "fixme",
	}

	lowerCred := strings.ToLower(credential)
	for _, placeholder := range placeholders {
		if strings.Contains(lowerCred, placeholder) {
			return &CredentialError{
				Type:    "placeholder",
				Message: "credential appears to be a placeholder value",
			}
		}
	}

	// Basic length validation (most API keys are at least 16 chars)
	if len(credential) < 8 {
		return &CredentialError{
			Type:    "too_short",
			Message: "credential appears too short to be valid",
		}
	}

	return nil
}

// CredentialError represents a credential validation error
type CredentialError struct {
	Type    string
	Message string
}

func (e *CredentialError) Error() string {
	return e.Message
}
