package templates

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/invopop/jsonschema"
)

type CommitStyle string

//go:embed styles/*.tmpl
var styles embed.FS

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
	Scope   string `json:"scope" jsonschema_description:"The area of the code affected by the commit"`
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

// CompileTemplate compiles a specific commit template at runtime
func (tm *TemplateManager) CompileTemplate(templateType CommitStyle) (CommitTemplate, error) {
	// Create a template with function map
	funcMap := template.FuncMap{
		// Add any functions you need here
		"secrets": func() string { return "" }, // Or implement proper secrets handling
	}

	commonFormat := `
Analyze the following git diff and generate a {{.Type}} commit message:

'''
{{.Diff}}
'''

The commit message should follow this format:
{{.Format}}

Where:

{{.Details}}

Please generate a commit message following this format{{.Extra}}.

The response should be a valid JSON object matching this schema:
{{.Schema}}
`

	createTemplate := func(name, typ, format, details, extra string, schema any) (CommitTemplate, error) {
		// Create template with function map
		tmpl, err := template.New(name).Funcs(funcMap).Parse(commonFormat)
		if err != nil {
			slog.Error("Failed to parse template", "error", err)
			return CommitTemplate{}, err
		}
		// Pass dynamic data into the template
		data := map[string]interface{}{
			"Type":    typ,
			"Diff":    tm.diff,
			"Format":  format,
			"Details": details,
			"Extra":   extra,
			"Schema":  schema,
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		if err != nil {
			slog.Error("Failed to execute template", "error", err)
			return CommitTemplate{}, err
		}
		// Create final template with function map as well
		finalTemplate, err := template.New(name + "_final").Funcs(funcMap).Parse(buf.String())
		if err != nil {
			slog.Error("Failed to parse template", "error", err)
			return CommitTemplate{}, err
		}
		return CommitTemplate{
				Template: finalTemplate,
				Schema:   jsonschema.Reflect(schema),
			}, nil
	}

	switch templateType {
	case "default":
		return createTemplate(
			"default",
			"",
			"<type>(<scope>): <subject>\n\n<body>\n\n<footer>",
			`- <type> is one of: feat, fix, docs, style, refactor, test, chore
- <scope> is optional and represents the module affected; generally small (deps, ci, etc)
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
			"",
			ConventionalCommit{},
		)
	case "conventional":
		return createTemplate(
			"conventional",
			"conventional",
			"<type>[optional scope]: <description>\n\n[optional body]\n\n[optional footer(s)]",
			`- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected; generally small (deps, ci, etc)
- <description> is a short summary in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
			"",
			ConventionalCommit{},
		)
	case "gitmojis":
		return createTemplate(
			"gitmojis",
			"gitmoji",
			"<gitmoji> <type>[optional scope]: <subject>\n\n<body>\n\n<footer>",
			`- <gitmoji> is an appropriate emoji for the change (e.g., 🐛 for bug fixes, ✨ for new features)
- <type> is one of: feat (✨), fix (🐛), docs (📝), style (💄), refactor (♻️), test (✅), chore (🔧), etc.
- <scope> is optional and represents the module affected; generally small (deps, ci, etc)
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
			", choosing an appropriate gitmoji",
			GitmojiCommitSchema{},
		)
	default:
		slog.Error("Unknown template type", "type", templateType)
		return CommitTemplate{}, fmt.Errorf("unknown template type: %s", templateType)
	}
}

var CommitStyleTemplateSchema = GenerateSchema[ConventionalCommit]()

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
