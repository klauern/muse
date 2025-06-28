package templates

import (
	"sync"
	"text/template"

	"github.com/invopop/jsonschema"
)

// TemplateRegistry provides thread-safe caching of compiled templates and schemas
type TemplateRegistry struct {
	templates map[string]*template.Template
	schemas   map[string]*jsonschema.Schema
	mutex     sync.RWMutex
}

// Global registry instance
var registry = &TemplateRegistry{
	templates: make(map[string]*template.Template),
	schemas:   make(map[string]*jsonschema.Schema),
}

// Get retrieves a cached template and schema if they exist
func (r *TemplateRegistry) Get(style string) (*template.Template, *jsonschema.Schema, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tmpl, tmplExists := r.templates[style]
	schema, schemaExists := r.schemas[style]

	return tmpl, schema, tmplExists && schemaExists
}

// Set stores a template and schema in the cache
func (r *TemplateRegistry) Set(style string, tmpl *template.Template, schema *jsonschema.Schema) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.templates[style] = tmpl
	r.schemas[style] = schema
}

// Clear removes all cached templates (useful for testing)
func (r *TemplateRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.templates = make(map[string]*template.Template)
	r.schemas = make(map[string]*jsonschema.Schema)
}

// GetRegistry returns the global registry instance
func GetRegistry() *TemplateRegistry {
	return registry
}
