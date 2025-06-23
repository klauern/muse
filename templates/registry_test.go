package templates

import (
	"sync"
	"testing"
	"text/template"

	"github.com/invopop/jsonschema"
)

func TestTemplateRegistry_Basic(t *testing.T) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	// Test empty registry
	_, _, exists := registry.Get("test")
	if exists {
		t.Error("should not find template in empty registry")
	}

	// Create test template and schema
	tmpl := template.New("test")
	schema := &jsonschema.Schema{Title: "test"}

	// Test set and get
	registry.Set("test", tmpl, schema)
	
	gotTmpl, gotSchema, exists := registry.Get("test")
	if !exists {
		t.Error("should find template after setting")
	}
	
	if gotTmpl != tmpl {
		t.Error("should return the same template instance")
	}
	
	if gotSchema != schema {
		t.Error("should return the same schema instance")
	}
}

func TestTemplateRegistry_ThreadSafety(t *testing.T) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	
	// Concurrent writes
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := "template_" + string(rune(id))
				tmpl := template.New(key)
				schema := &jsonschema.Schema{Title: key}
				registry.Set(key, tmpl, schema)
			}
		}(i)
	}

	// Concurrent reads
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := "template_" + string(rune(id))
				registry.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify we can still access the registry
	tmpl := template.New("final_test")
	schema := &jsonschema.Schema{Title: "final_test"}
	registry.Set("final_test", tmpl, schema)
	
	gotTmpl, gotSchema, exists := registry.Get("final_test")
	if !exists || gotTmpl != tmpl || gotSchema != schema {
		t.Error("registry should be functional after concurrent access")
	}
}

func TestTemplateRegistry_Clear(t *testing.T) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	// Add some templates
	tmpl1 := template.New("test1")
	schema1 := &jsonschema.Schema{Title: "test1"}
	registry.Set("test1", tmpl1, schema1)

	tmpl2 := template.New("test2")
	schema2 := &jsonschema.Schema{Title: "test2"}
	registry.Set("test2", tmpl2, schema2)

	// Verify they exist
	_, _, exists1 := registry.Get("test1")
	_, _, exists2 := registry.Get("test2")
	if !exists1 || !exists2 {
		t.Error("templates should exist before clear")
	}

	// Clear registry
	registry.Clear()

	// Verify they're gone
	_, _, exists1 = registry.Get("test1")
	_, _, exists2 = registry.Get("test2")
	if exists1 || exists2 {
		t.Error("templates should not exist after clear")
	}
}

func TestGetRegistry_Singleton(t *testing.T) {
	// Should return the same instance
	registry1 := GetRegistry()
	registry2 := GetRegistry() 

	if registry1 != registry2 {
		t.Error("GetRegistry should return the same singleton instance")
	}

	// Test that it's functional
	tmpl := template.New("singleton_test")
	schema := &jsonschema.Schema{Title: "singleton_test"}
	
	registry1.Set("singleton_test", tmpl, schema)
	
	gotTmpl, gotSchema, exists := registry2.Get("singleton_test")
	if !exists || gotTmpl != tmpl || gotSchema != schema {
		t.Error("global registry should be shared across calls")
	}
}

func TestTemplateRegistry_PartialData(t *testing.T) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	// Add only template, no schema
	tmpl := template.New("partial")
	registry.templates["partial"] = tmpl

	// Should not exist because schema is missing
	_, _, exists := registry.Get("partial")
	if exists {
		t.Error("should not return template without schema")
	}

	// Add only schema, no template
	schema := &jsonschema.Schema{Title: "partial"}
	registry.schemas["partial2"] = schema

	// Should not exist because template is missing
	_, _, exists = registry.Get("partial2")
	if exists {
		t.Error("should not return schema without template")
	}

	// Add both - should work
	registry.Set("complete", tmpl, schema)
	_, _, exists = registry.Get("complete")
	if !exists {
		t.Error("should return complete template+schema pair")
	}
}

func BenchmarkTemplateRegistry_Get(b *testing.B) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	// Pre-populate registry
	for i := 0; i < 100; i++ {
		key := "template_" + string(rune(i))
		tmpl := template.New(key)
		schema := &jsonschema.Schema{Title: key}
		registry.Set(key, tmpl, schema)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry.Get("template_50")
		}
	})
}

func BenchmarkTemplateRegistry_Set(b *testing.B) {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "template_" + string(rune(i))
			tmpl := template.New(key)
			schema := &jsonschema.Schema{Title: key}
			registry.Set(key, tmpl, schema)
			i++
		}
	})
}