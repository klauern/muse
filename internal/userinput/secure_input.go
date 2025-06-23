package userinput

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// SecureInputHandler provides secure user input handling with timeout and validation
type SecureInputHandler struct {
	timeout time.Duration
	maxLen  int
}

// InputValidationError represents an error in user input validation
type InputValidationError struct {
	Input  string
	Reason string
}

func (e InputValidationError) Error() string {
	return fmt.Sprintf("invalid input: %s (input: %q)", e.Reason, e.Input)
}

// NewSecureInputHandler creates a new secure input handler with default settings
func NewSecureInputHandler() *SecureInputHandler {
	return &SecureInputHandler{
		timeout: 30 * time.Second, // 30 second timeout
		maxLen:  100,              // Maximum input length
	}
}

// SetTimeout configures the input timeout
func (s *SecureInputHandler) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// SetMaxLength configures the maximum input length
func (s *SecureInputHandler) SetMaxLength(maxLen int) {
	s.maxLen = maxLen
}

// PromptYesNo prompts the user for a yes/no response with timeout and validation
func (s *SecureInputHandler) PromptYesNo(ctx context.Context, prompt string) (bool, error) {
	fmt.Print(prompt)

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Channel to receive input
	inputChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start input reading in goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := scanner.Text()
			inputChan <- input
		} else {
			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("failed to read input: %w", err)
			} else {
				errChan <- fmt.Errorf("no input received")
			}
		}
	}()

	// Wait for input or timeout
	select {
	case input := <-inputChan:
		return s.validateYesNoInput(input)
	case err := <-errChan:
		return false, err
	case <-timeoutCtx.Done():
		return false, fmt.Errorf("input timeout after %v", s.timeout)
	}
}

// validateYesNoInput validates and sanitizes yes/no input
func (s *SecureInputHandler) validateYesNoInput(input string) (bool, error) {
	// Sanitize input
	sanitized := s.sanitizeInput(input)

	// Validate length
	if len(sanitized) > s.maxLen {
		truncated := sanitized
		if len(truncated) > 50 {
			truncated = truncated[:50] + "..."
		}
		return false, InputValidationError{
			Input:  truncated,
			Reason: fmt.Sprintf("input too long (max %d chars)", s.maxLen),
		}
	}

	// Normalize to lowercase for comparison
	normalized := strings.ToLower(strings.TrimSpace(sanitized))

	// Check for valid yes/no responses
	switch normalized {
	case "y", "yes", "true", "1":
		return true, nil
	case "n", "no", "false", "0":
		return false, nil
	case "":
		return false, InputValidationError{
			Input:  "",
			Reason: "empty input not allowed",
		}
	default:
		return false, InputValidationError{
			Input:  normalized,
			Reason: "must be 'y', 'yes', 'n', or 'no'",
		}
	}
}

// sanitizeInput removes potentially dangerous characters and patterns
func (s *SecureInputHandler) sanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	var sanitized strings.Builder
	for _, r := range input {
		// Allow printable characters, space, newline, tab
		if r >= 32 && r < 127 || r == '\n' || r == '\t' {
			sanitized.WriteRune(r)
		}
		// Skip other control characters
	}

	result := sanitized.String()

	// Remove excessive whitespace
	result = strings.TrimSpace(result)

	// Collapse multiple whitespace into single spaces
	words := strings.Fields(result)
	result = strings.Join(words, " ")

	return result
}

// PromptWithValidation prompts for input with custom validation
func (s *SecureInputHandler) PromptWithValidation(ctx context.Context, prompt string, validator func(string) error) (string, error) {
	fmt.Print(prompt)

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Channel to receive input
	inputChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start input reading in goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := scanner.Text()
			inputChan <- input
		} else {
			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("failed to read input: %w", err)
			} else {
				errChan <- fmt.Errorf("no input received")
			}
		}
	}()

	// Wait for input or timeout
	select {
	case input := <-inputChan:
		// Sanitize input
		sanitized := s.sanitizeInput(input)

		// Validate length
		if len(sanitized) > s.maxLen {
			truncated := sanitized
			if len(truncated) > 50 {
				truncated = truncated[:50] + "..."
			}
			return "", InputValidationError{
				Input:  truncated,
				Reason: fmt.Sprintf("input too long (max %d chars)", s.maxLen),
			}
		}

		// Run custom validation
		if validator != nil {
			if err := validator(sanitized); err != nil {
				return "", fmt.Errorf("validation failed: %w", err)
			}
		}

		return sanitized, nil
	case err := <-errChan:
		return "", err
	case <-timeoutCtx.Done():
		return "", fmt.Errorf("input timeout after %v", s.timeout)
	}
}
