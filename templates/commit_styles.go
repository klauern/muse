package templates

import (
	"fmt"
	"text/template"

	"github.com/invopop/jsonschema"
)

const (
	ConventionalCommitStyle = "conventional"
	GitmojiCommitStyle      = "gitmoji"
)

// CommitTemplate represents a template for generating commit messages
type CommitTemplate struct {
	Template *template.Template
	Schema   *jsonschema.Schema
}

// ConventionalCommit represents the structure of a conventional commit
type ConventionalCommit struct {
	Type    string `json:"type" jsonschema:"enum=feat,enum=fix,enum=chore,enum=docs,enum=style,enum=refactor,enum=test,enum=build,enum=ci,enum=perf,enum=revert,description=Type of commit following conventional commits"`
	Scope   string `json:"scope,omitempty" jsonschema:"pattern=^[a-zA-Z0-9-_]+$,description=The area of the code affected by the commit,default="`
	Subject string `json:"subject" jsonschema:"minLength=5,maxLength=72,description=A short summary of the change"`
	Body    string `json:"body,omitempty" jsonschema:"description=A detailed description of the change"`
	Footer  string `json:"footer,omitempty" jsonschema:"pattern=^(Closes|Fixes) #[0-9]+$,description=Any issue references or breaking change notes"`
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
	DefaultCommit      CommitTemplate
	ConventionalCommit CommitTemplate
	Gitmojis           CommitTemplate
}

// NewTemplateManager creates and returns a new TemplateManager
func NewTemplateManager() (*TemplateManager, error) {
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

	createTemplate := func(name, typ, format, details, extra string, schema any) (CommitTemplate, error) {
		tmpl, err := template.New(name).Parse(commonFormat)
		if err != nil {
			return CommitTemplate{}, err
		}
		return CommitTemplate{
			Template: tmpl.Option("missingkey=error"),
			Schema:   jsonschema.Reflect(schema),
		}, nil
	}

	defaultCommit, err := createTemplate(
		"default",
		"",
		"<type>(<scope>): <subject>\n\n<body>\n\n<footer>",
		`- <type> is one of: feat, fix, docs, style, refactor, test, chore
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		"",
		ConventionalCommit{},
	)
	if err != nil {
		return nil, err
	}

	conventionalCommit, err := createTemplate(
		"conventional",
		"conventional",
		"<type>[optional scope]: <description>\n\n[optional body]\n\n[optional footer(s)]",
		`- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected
- <description> is a short summary in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		"",
		ConventionalCommit{},
	)
	if err != nil {
		return nil, err
	}

	gitmojis, err := createTemplate(
		"gitmojis",
		"gitmoji",
		"<gitmoji> <type>[optional scope]: <subject>\n\n<body>\n\n<footer>",
		`- <gitmoji> is an appropriate emoji for the change (e.g., 🐛 for bug fixes, ✨ for new features)
- <type> is one of: feat (✨), fix (🐛), docs (📝), style (💄), refactor (♻️), test (✅), chore (🔧)
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)`,
		", choosing an appropriate gitmoji",
		GitmojiCommitSchema{},
	)
	if err != nil {
		return nil, err
	}

	return &TemplateManager{
		DefaultCommit:      defaultCommit,
		ConventionalCommit: conventionalCommit,
		Gitmojis:           gitmojis,
	}, nil
}
