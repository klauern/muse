package templates

import (
	"text/template"

	"github.com/sashabaranov/go-openai/jsonschema"
)

type CommitTemplate struct {
	Template *template.Template
	Schema   jsonschema.Definition
}

var (
	DefaultCommitTemplate      CommitTemplate
	ConventionalCommitTemplate CommitTemplate
	GitmojisTemplate           CommitTemplate
)

func init() {
	commonFormat := `
Generate a {{.Type}} commit message for the following git diff:
{{.Diff}}

{{if .Context}}Additional context:
{{.Context}}
{{end}}

The commit message should follow this format:
{{.Format}}
Where:
{{.Details}}

Please generate a commit message following this format{{.Extra}}.

The response should be a valid JSON object matching this schema:
{{.Schema}}
`

	createTemplate := func(name, typ, format, details, extra string, schema jsonschema.Definition) CommitTemplate {
		return CommitTemplate{
			Template: template.Must(template.New(name).Parse(commonFormat)).Option("missingkey=error"),
			Schema:   schema,
		}
	}

	commonSchema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"type": {
				Type: jsonschema.String,
				Enum: []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
			},
			"scope": {
				Type:        jsonschema.String,
				Description: "optional and represents the module affected",
			},
			"subject": {
				Type:        jsonschema.String,
				Description: "a short description",
			},
			"body": {
				Type:        jsonschema.String,
				Description: "provides additional context (optional)",
			},
			"footer": {
				Type:        jsonschema.String,
				Description: "mentions any breaking changes or closed issues (optional)",
			},
		},
		Required: []string{"type", "subject"},
	}

	DefaultCommitTemplate = createTemplate(
		"default",
		"",
		"<type>(<scope>): <subject>\n\n<body>\n\n<footer>",
		`- <type> is one of: feat, fix, docs, style, refactor, test, chore
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		"",
		commonSchema,
	)

	ConventionalCommitTemplate = createTemplate(
		"conventional",
		"conventional",
		"<type>[optional scope]: <description>\n\n[optional body]\n\n[optional footer(s)]",
		`- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected
- <description> is a short summary in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		"",
		commonSchema,
	)

	gitmojisSchema := commonSchema
	gitmojisSchema.Properties["gitmoji"] = jsonschema.Definition{
		Type:        jsonschema.String,
		Description: "an appropriate emoji for the change",
	}
	gitmojisSchema.Required = append(gitmojisSchema.Required, "gitmoji")

	GitmojisTemplate = createTemplate(
		"gitmojis",
		"gitmoji",
		"<gitmoji> <type>[optional scope]: <subject>\n\n<body>\n\n<footer>",
		`- <gitmoji> is an appropriate emoji for the change (e.g., 🐛 for bug fixes, ✨ for new features)
- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		", choosing an appropriate gitmoji",
		gitmojisSchema,
	)
}
