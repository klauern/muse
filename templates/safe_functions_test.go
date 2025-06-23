package templates

import (
	"strings"
	"testing"
	"text/template"
)

func TestSafeFuncMap_NoUnsafeFunctions(t *testing.T) {
	funcMap := SafeFuncMap()

	// Ensure dangerous functions are not present
	unsafeFuncs := []string{
		"lookPath", "stat", "abs", "exec", "eval", "system",
		"open", "create", "mkdir", "remove", "chmod",
	}

	for _, unsafeFunc := range unsafeFuncs {
		if _, exists := funcMap[unsafeFunc]; exists {
			t.Errorf("Unsafe function %s found in function map", unsafeFunc)
		}
	}
}

func TestSafeFuncMap_OnlySafeFunctions(t *testing.T) {
	funcMap := SafeFuncMap()

	// Verify expected safe functions are present
	expectedFuncs := []string{
		"upper", "lower", "title", "trim", "contains",
		"hasPrefix", "hasSuffix", "basename", "extname",
		"cleanPath", "quote", "sanitize",
	}

	for _, expectedFunc := range expectedFuncs {
		if _, exists := funcMap[expectedFunc]; !exists {
			t.Errorf("Expected safe function %s not found in function map", expectedFunc)
		}
	}
}

func TestSanitizeTemplateInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "template delimiters",
			input:    "{{.Malicious}}",
			expected: "&#123;&#123;.Malicious&#125;&#125;",
		},
		{
			name:     "script tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert('xss')&lt;/script&gt;",
		},
		{
			name:     "html entities",
			input:    "<div>test & stuff</div>",
			expected: "&lt;div&gt;test &amp; stuff&lt;/div&gt;",
		},
		{
			name:     "normal text",
			input:    "This is normal text",
			expected: "This is normal text",
		},
		{
			name:     "length limit",
			input:    strings.Repeat("a", 60000),
			expected: strings.Repeat("a", 50000) + "... [truncated for security]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTemplateInput(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeTemplateInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanPath_PreventDirectoryTraversal(t *testing.T) {
	funcMap := SafeFuncMap()
	cleanPathFunc := funcMap["cleanPath"].(func(string) string)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal path",
			input:    "templates/style.tmpl",
			expected: "templates/style.tmpl",
		},
		{
			name:     "directory traversal attempt",
			input:    "../../../etc/passwd",
			expected: "", // Should be blocked
		},
		{
			name:     "hidden directory traversal",
			input:    "templates/../../../etc/passwd",
			expected: "", // Should be blocked
		},
		{
			name:     "clean relative path",
			input:    "./templates/./style.tmpl",
			expected: "templates/style.tmpl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPathFunc(tt.input)
			if result != tt.expected {
				t.Errorf("cleanPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTemplateExecution_PreventCodeInjection(t *testing.T) {
	funcMap := SafeFuncMap()

	// Test template with potentially malicious input
	tmplText := `Hello {{.Name | sanitize}}`
	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]string{
		"Name": "{{.Secret}} <script>alert('xss')</script>",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	expected := "Hello &#123;&#123;.Secret&#125;&#125; &lt;script&gt;alert('xss')&lt;/script&gt;"

	if result != expected {
		t.Errorf("Template execution result = %v, want %v", result, expected)
	}

	// Ensure no actual template injection occurred
	if strings.Contains(result, "{{") && !strings.Contains(result, "&#123;") {
		t.Error("Template injection vulnerability detected")
	}
}

func TestSubstr_SafeBounds(t *testing.T) {
	funcMap := SafeFuncMap()
	substrFunc := funcMap["substr"].(func(string, int, int) string)

	tests := []struct {
		name     string
		str      string
		start    int
		length   int
		expected string
	}{
		{
			name:     "normal substring",
			str:      "hello world",
			start:    0,
			length:   5,
			expected: "hello",
		},
		{
			name:     "out of bounds start",
			str:      "hello",
			start:    10,
			length:   5,
			expected: "",
		},
		{
			name:     "negative start",
			str:      "hello",
			start:    -1,
			length:   5,
			expected: "",
		},
		{
			name:     "length beyond string",
			str:      "hello",
			start:    2,
			length:   10,
			expected: "llo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substrFunc(tt.str, tt.start, tt.length)
			if result != tt.expected {
				t.Errorf("substr() = %v, want %v", result, tt.expected)
			}
		})
	}
}
