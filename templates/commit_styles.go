package templates

import "text/template"

var (
	DefaultCommitTemplate      *template.Template
	ConventionalCommitTemplate *template.Template
	GitmojisTemplate           *template.Template
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
`

	createTemplate := func(name, typ, format, details, extra string) *template.Template {
		return template.Must(template.New(name).Parse(commonFormat))
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
	)

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
	)
}
