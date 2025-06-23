package templates

import (
	"strings"
	"testing"
)

func TestTemplateManager_CompileTemplate(t *testing.T) {
	// Clear registry before tests
	GetRegistry().Clear()
	
	tests := []struct {
		name         string
		style        CommitStyle
		diff         string
		expectError  bool
		expectCache  bool
	}{
		{
			name:        "conventional template",
			style:       "conventional",
			diff:        "diff --git a/test.go b/test.go\n+added line",
			expectError: false,
			expectCache: true,
		},
		{
			name:        "gitmoji template",
			style:       "gitmoji",
			diff:        "diff --git a/test.go b/test.go\n+added line",
			expectError: false,
			expectCache: true,
		},
		{
			name:        "default template",
			style:       "default",
			diff:        "diff --git a/test.go b/test.go\n+added line",
			expectError: false,
			expectCache: true,
		},
		{
			name:        "unknown template",
			style:       "unknown",
			diff:        "diff --git a/test.go b/test.go\n+added line",
			expectError: true,
			expectCache: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTemplateManager(tt.diff, tt.style)
			
			// First compilation
			result, err := tm.CompileTemplate(tt.style)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for style %s, but got none", tt.style)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if result.Template == nil {
				t.Error("template should not be nil")
			}
			
			if result.Schema == nil {
				t.Error("schema should not be nil")
			}
			
			// Test caching - second compilation should use cache
			if tt.expectCache {
				result2, err2 := tm.CompileTemplate(tt.style)
				if err2 != nil {
					t.Fatalf("unexpected error on cached template: %v", err2)
				}
				
				// Should be the same template instance (from cache)
				if result.Template != result2.Template {
					t.Error("expected cached template to be the same instance")
				}
			}
		})
	}
}

func TestTemplateData_Security(t *testing.T) {
	maliciousDiff := `{{.Secret}}{{lookPath "evil"}}{{stat "/etc/passwd"}}{{abs "/tmp"}}`
	
	tm := NewTemplateManager(maliciousDiff, "conventional")
	data := tm.GetTemplateData()
	
	sanitizedDiff, ok := data["Diff"].(string)
	if !ok {
		t.Fatal("Diff should be a string")
	}
	
	// Should escape template delimiters
	if strings.Contains(sanitizedDiff, "{{") || strings.Contains(sanitizedDiff, "}}") {
		t.Error("template delimiters should be escaped")
	}
	
	// Should contain escaped versions
	if !strings.Contains(sanitizedDiff, "&#123;&#123;") {
		t.Error("template delimiters should be escaped to HTML entities")
	}
}

func TestGenerateSchemaForStyle(t *testing.T) {
	tm := NewTemplateManager("test diff", "conventional")
	
	tests := []struct {
		name  string
		style CommitStyle
	}{
		{
			name:  "conventional schema",
			style: "conventional",
		},
		{
			name:  "default schema",
			style: "default",
		},
		{
			name:  "gitmoji schema",
			style: "gitmoji",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tm.generateSchemaForStyle(tt.style)
			
			if schema == nil {
				t.Error("schema should not be nil")
			}
			
			// Schema should have required properties
			if schema.Properties == nil {
				t.Error("schema should have properties")
			}
		})
	}
}

func TestTemplateExecution_WithRealData(t *testing.T) {
	// Clear registry before test
	GetRegistry().Clear()
	
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
 
+import "fmt"
 func main() {}`
	
	tm := NewTemplateManager(diff, "conventional")
	
	result, err := tm.CompileTemplate("conventional")
	if err != nil {
		t.Fatalf("failed to compile template: %v", err)
	}
	
	// Execute template with real data
	var buf strings.Builder
	data := tm.GetTemplateData()
	
	err = result.Template.Execute(&buf, data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
	
	output := buf.String()
	
	// Should contain the diff
	if !strings.Contains(output, "package main") {
		t.Error("output should contain diff content")
	}
	
	// Should contain schema information
	if !strings.Contains(output, "schema") {
		t.Error("output should contain schema information")
	}
}

func TestTemplateManager_ConcurrentAccess(t *testing.T) {
	// Clear registry before test
	GetRegistry().Clear()
	
	// Test concurrent access to template manager
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			tm := NewTemplateManager("test diff", "conventional")
			_, err := tm.CompileTemplate("conventional")
			if err != nil {
				t.Errorf("concurrent template compilation failed: %v", err)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify template is cached
	if tmpl, schema, exists := GetRegistry().Get("conventional"); !exists {
		t.Error("template should be cached after concurrent access")
	} else {
		if tmpl == nil || schema == nil {
			t.Error("cached template and schema should not be nil")
		}
	}
}

func TestReadTemplateFile(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "conventional template exists",
			path:        "styles/conventional.tmpl",
			expectError: false,
		},
		{
			name:        "gitmoji template exists",
			path:        "styles/gitmoji.tmpl",
			expectError: false,
		},
		{
			name:        "default template exists", 
			path:        "styles/default.tmpl",
			expectError: false,
		},
		{
			name:        "nonexistent template",
			path:        "styles/nonexistent.tmpl",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := ReadTemplateFile(tt.path)
			
			if tt.expectError {
				if err == nil {
					t.Error("expected error for nonexistent file")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if content == "" {
				t.Error("template content should not be empty")
			}
			
			// Should contain template placeholders
			if !strings.Contains(content, "{{.Diff}}") {
				t.Error("template should contain {{.Diff}} placeholder")
			}
			
			if !strings.Contains(content, "{{.Schema}}") {
				t.Error("template should contain {{.Schema}} placeholder")
			}
		})
	}
}

func TestListTemplateFiles(t *testing.T) {
	files, err := ListTemplateFiles()
	if err != nil {
		t.Fatalf("failed to list template files: %v", err)
	}
	
	expectedFiles := []string{"conventional.tmpl", "default.tmpl", "gitmoji.tmpl"}
	
	if len(files) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(files))
	}
	
	for _, expectedFile := range expectedFiles {
		found := false
		for _, file := range files {
			if file == expectedFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s not found in list", expectedFile)
		}
	}
}