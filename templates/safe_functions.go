package templates

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// SafeFuncMap returns a whitelist of safe template functions
// that cannot be exploited for file system access or command execution
func SafeFuncMap() template.FuncMap {
	return template.FuncMap{
		// String operations (safe)
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"replace":   strings.ReplaceAll,
		"split":     strings.Split,
		"join":      strings.Join,

		// Safe path operations (no filesystem access)
		"basename": filepath.Base,
		"extname":  filepath.Ext,
		"cleanPath": func(path string) string {
			// Only clean relative paths, don't resolve absolute paths
			cleaned := filepath.Clean(path)
			// Prevent directory traversal by blocking paths with ".."
			if strings.Contains(cleaned, "..") {
				return ""
			}
			return cleaned
		},

		// Format helpers
		"quote": func(s string) string {
			return fmt.Sprintf("%q", s)
		},
		"sprintf": fmt.Sprintf,

		// Safe utility functions
		"len": func(s string) int {
			return len(s)
		},
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"default": func(defaultVal, val string) string {
			if val == "" {
				return defaultVal
			}
			return val
		},

		// Template data validation
		"sanitize": sanitizeTemplateInput,
	}
}

// sanitizeTemplateInput sanitizes user input to prevent template injection
func sanitizeTemplateInput(input string) string {
	// Escape HTML entities first to avoid double-escaping
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")

	// Escape template delimiters
	input = strings.ReplaceAll(input, "{{", "&#123;&#123;")
	input = strings.ReplaceAll(input, "}}", "&#125;&#125;")

	// Limit length to prevent memory exhaustion
	const maxInputLength = 50000 // 50KB limit
	if len(input) > maxInputLength {
		input = input[:maxInputLength] + "... [truncated for security]"
	}

	return input
}
