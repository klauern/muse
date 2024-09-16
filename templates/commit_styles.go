package templates

import "text/template"

var (
	DefaultCommitTemplate = template.Must(template.New("default").Parse(`
Generate a commit message for the following git diff:

{{.Diff}}

Additional context:
{{.Context}}

The commit message should follow this format:
<type>(<scope>): <subject>

<body>

<footer>

Where:
- <type> is one of: feat, fix, docs, style, refactor, test, chore
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)

Please generate a commit message following this format.
`))

	ConventionalCommitTemplate = template.Must(template.New("conventional").Parse(`
Generate a conventional commit message for the following git diff:

{{.Diff}}

Additional context:
{{.Context}}

The commit message should follow the Conventional Commits specification:
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]

Where:
- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected
- <description> is a short summary in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)

Please generate a commit message following this format.
`))

	GitmojisTemplate = template.Must(template.New("gitmojis").Parse(`
Generate a gitmoji commit message for the following git diff:

{{.Diff}}

Additional context:
{{.Context}}

The commit message should follow this format:
<gitmoji> <type>[optional scope]: <subject>

<body>

<footer>

Where:
- <gitmoji> is an appropriate emoji for the change (e.g., 🐛 for bug fixes, ✨ for new features)
- <type> is one of: feat, fix, docs, style, refactor, test, chore, etc.
- <scope> is optional and represents the module affected
- <subject> is a short description in the present tense
- <body> provides additional context (optional)
- <footer> mentions any breaking changes or closed issues (optional)

Please generate a commit message following this format, choosing an appropriate gitmoji.
`))
)
