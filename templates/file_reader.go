package templates

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed styles/*.tmpl
var templateFS embed.FS

// ReadTemplateFile reads a template file from the embedded filesystem
func ReadTemplateFile(path string) (string, error) {
	content, err := fs.ReadFile(templateFS, path)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", path, err)
	}
	return string(content), nil
}

// ListTemplateFiles returns a list of available template files
func ListTemplateFiles() ([]string, error) {
	entries, err := fs.ReadDir(templateFS, "styles")
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name()[len(entry.Name())-5:] == ".tmpl" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
