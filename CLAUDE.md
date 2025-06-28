# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Muse is an AI-powered Git commit message generator written in Go. It integrates with Git hooks to automatically generate meaningful commit messages using various LLM providers (OpenAI, Anthropic, Ollama) based on staged changes. The tool supports multiple commit message styles including Conventional Commits and Gitmoji formats.

## Common Commands

### Building and Development
```bash
# Build the binary
task build

# Install locally to $HOME/go/bin
task install

# Format code using gofumpt
task format

# Run all linters
task lint

# Run Go vet
task vet
```

### Testing
```bash
# Run all tests (unit + integration)
task test

# Run only unit tests
task test:unit

# Run integration tests (requires API keys)
task test:integration

# Run specific provider integration tests
task test:integration:openai
task test:integration:anthropic

# Generate and view coverage report
task test:cover

# Show coverage summary
task test:cover:show

# Test individual packages
task test:cmd
task test:config
task test:llm
task test:hooks
task test:templates
task test:memory
```

### Version Management
```bash
# Bump patch version (e.g., 1.0.0 -> 1.0.1)
task bump:patch

# Bump minor version (e.g., 1.0.0 -> 1.1.0)
task bump:minor

# Bump major version (e.g., 1.0.0 -> 2.0.0)
task bump:major
```

## Architecture Overview

### Core Components

**CLI Layer (`cmd/`)**: Built with urfave/cli/v2, provides commands for install, uninstall, configure, status, and prepare-commit-msg operations.

**Configuration (`config/`)**: Uses koanf library for configuration management with support for YAML files, environment variables, and embedded defaults. Configuration is loaded from:
- `./muse.yaml` (local directory)
- `$XDG_CONFIG_HOME/muse/muse.yaml`
- `$HOME/.config/muse/muse.yaml`

**LLM Providers (`llm/`)**: Plugin-based architecture using a provider registry pattern. Each provider implements the `LLMService` interface and registers itself via `init()` functions. Currently supports OpenAI with structured outputs using JSON Schema.

**Template System (`templates/`)**: Dynamic template compilation system that generates LLM prompts based on commit styles. Uses Go's `text/template` package with JSON Schema validation for structured outputs.

**Git Hooks (`hooks/`)**: Implements the prepare-commit-msg Git hook interface, handles diff extraction, message generation, and user interaction (preview/dry-run modes).

### Key Design Patterns

- **Provider Registry**: LLM providers self-register using `init()` functions
- **Template Compilation**: Templates are dynamically compiled at runtime with context-specific data
- **Structured Outputs**: Uses JSON Schema to ensure consistent LLM responses
- **Configuration Cascade**: Environment variables override YAML configuration

### Configuration Structure

The configuration supports multiple commit styles:
- `conventional`: Conventional Commits format
- `gitmojis`: Conventional Commits with emoji prefixes
- `default`: Simple semantic commit format

Each LLM provider has its own configuration block under `llm.config` with provider-specific settings like API keys, model selection, and API base URLs.

### Internal Architecture

**Internal Utilities (`internal/`)**: Core internal packages that provide foundational functionality:
- `security/`: Credential validation and masking to prevent API key exposure
- `git/`: Safe Git operations with proper error handling
- `userinput/`: Secure user input handling with validation
- `fileops/`: Atomic file operations for safe config management

**Memory Bank (`memory-bank/`)**: Project documentation and context storage for maintaining project history and decision records.

### Key Dependencies

- **CLI Framework**: `urfave/cli/v2` for command-line interface
- **Configuration**: `koanf/koanf` for configuration management with YAML and env support
- **LLM Integration**: `openai/openai-go` for OpenAI API integration with structured outputs
- **JSON Schema**: `invopop/jsonschema` for structured output validation
- **UI Elements**: `briandowns/spinner` for loading indicators during API calls

### Important Development Notes

- Integration tests require valid API keys set as environment variables
- The `templates` package uses JSON Schema reflection for OpenAI structured outputs
- Git diff analysis happens via `git diff --cached` to analyze staged changes
- The hook system supports preview mode for user approval and dry-run mode for testing
- Thread-safe template caching is implemented in the template registry
- Never use `testify/assert` in tests (use standard Go testing patterns)

## Development Best Practices

- Always update memory-bank/md files when implementing changes
- Use the provider registry pattern when adding new LLM providers
- Ensure all credential handling goes through the security package
- Template changes should include JSON Schema validation updates