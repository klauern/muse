package templates

import (
	"fmt"
	"text/template"

	"github.com/invopop/jsonschema"
)

type CommitStyle string

const (
	ConventionalCommitStyle CommitStyle = "conventional"
	GitmojiCommitStyle      CommitStyle = "gitmoji"
)

// CommitTemplate represents a template for generating commit messages
type CommitTemplate struct {
	Template *template.Template
	Schema   *jsonschema.Schema
}

// ConventionalCommit represents the structure of a conventional commit
type ConventionalCommit struct {
	Type    string `json:"type" jsonschema:"enum=feat,enum=fix,enum=chore,enum=docs,enum=style,enum=refactor,enum=test,enum=build,enum=ci,enum=perf,enum=revert" jsonschema_description:"Type of commit following conventional commits"`
	Scope   string `json:"scope" jsonschema_description:"The area/section of the code affected by the commit.  Usually short (tests, deps, ci, etc)"`
	Subject string `json:"subject" jsonschema_description:"A short summary (5 to 72 characters) of the change"`
	Body    string `json:"body" jsonschema_description:"A detailed description of the change"`
	Footer  string `json:"footer" jsonschema_description:"Any issue references or breaking change notes, generally in the format of Closes-#123 or Fixes-#123"`
}

func (c *ConventionalCommit) String() string {
	return fmt.Sprintf("%s(%s): %s\n\n%s", c.Type, c.Scope, c.Subject, c.Body)
}

// GitmojiCommitSchema extends CommitSchema with a gitmoji field
type GitmojiCommitSchema struct {
	ConventionalCommit
	Gitmoji string `json:"gitmoji" jsonschema:"description=an appropriate emoji for the change"`
}

// TemplateManager manages different commit templates
type TemplateManager struct {
	diff  string
	style CommitStyle
}

// NewTemplateManager creates and returns a new TemplateManager
func NewTemplateManager(diff string, style CommitStyle) *TemplateManager {
	return &TemplateManager{
		diff:  diff,
		style: style,
	}
}

// CompileTemplate compiles a specific commit template using single-pass compilation with caching
func (tm *TemplateManager) CompileTemplate(templateType CommitStyle) (CommitTemplate, error) {
	// Check cache first
	if tmpl, schema, exists := GetRegistry().Get(string(templateType)); exists {
		return CommitTemplate{Template: tmpl, Schema: schema}, nil
	}

	// Load template from file (not hardcoded strings)
	templateFile := fmt.Sprintf("styles/%s.tmpl", templateType)
	templateContent, err := ReadTemplateFile(templateFile)
	if err != nil {
		return CommitTemplate{}, fmt.Errorf("failed to read template file %s: %w", templateFile, err)
	}

	// Single compilation with safe function map
	tmpl, err := template.New(string(templateType)).
		Funcs(SafeFuncMap()).
		Parse(templateContent)
	if err != nil {
		return CommitTemplate{}, fmt.Errorf("failed to parse template: %w", err)
	}

	// Generate schema for the commit style
	schema := tm.generateSchemaForStyle(templateType)

	// Cache for future use
	GetRegistry().Set(string(templateType), tmpl, schema)

	return CommitTemplate{Template: tmpl, Schema: schema}, nil
}

// generateSchemaForStyle generates the appropriate schema for a commit style
func (tm *TemplateManager) generateSchemaForStyle(style CommitStyle) *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	switch style {
	case "gitmoji", "gitmojis":
		return reflector.Reflect(GitmojiCommitSchema{})
	default:
		return reflector.Reflect(ConventionalCommit{})
	}
}

// GetTemplateData prepares the data for template execution
func (tm *TemplateManager) GetTemplateData() map[string]interface{} {
	// Sanitize diff input to prevent template injection
	sanitizedDiff := sanitizeTemplateInput(tm.diff)

	// Generate schema for the current style
	schema := tm.generateSchemaForStyle(tm.style)

	return map[string]interface{}{
		"Diff":   sanitizedDiff,
		"Schema": schema,
	}
}

// GenerateSchema generates a JSON schema for a given type, adhering to the subset of JSON Schema supported by OpenAI's structured outputs.
func GenerateSchema[T any]() any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}
