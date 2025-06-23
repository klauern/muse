# techContext.md

**Purpose:**
Documents technologies used, development setup, technical constraints, dependencies, and tool usage patterns.

---

## Technologies Used

### Core Technologies

- **Go 1.24.4**: Primary programming language for CLI application
- **Git Hooks**: prepare-commit-msg hook for Git workflow integration
- **YAML**: Configuration file format with human-readable syntax
- **JSON Schema**: Structured validation for LLM responses

### Key Dependencies

#### CLI and Configuration

- **github.com/urfave/cli/v2 v2.27.4**: Command-line interface framework
- **github.com/knadh/koanf v1.5.0**: Configuration management with multiple sources
- **github.com/invopop/jsonschema v0.12.0**: JSON schema generation and validation

#### LLM Integration

- **github.com/openai/openai-go v0.1.0-alpha.25**: OpenAI API client (official Go SDK)
- Support planned for:
  - Anthropic Claude API
  - Ollama local models
  - Other LLM providers via plugin architecture

#### User Experience

- **github.com/briandowns/spinner v1.23.1**: CLI loading indicators
- **github.com/fatih/color**: Terminal color output
- **golang.org/x/term**: Terminal interaction handling

#### Utilities

- **gopkg.in/yaml.v3**: YAML parsing and generation
- **github.com/mitchellh/mapstructure v1.5.0**: Configuration structure mapping

### Build and Development Tools

- **Go Modules**: Dependency management
- **GoReleaser**: Cross-platform binary distribution
- **Taskfile**: Build automation and task management
- **mise**: Development environment management

## Development Setup

### Prerequisites

- **Go 1.24.4+**: Latest Go version with modern features
- **Git**: Version control system with hook support
- **make** or **task**: Build automation (optional)

### Environment Setup

```bash
# Clone repository
git clone https://github.com/klauern/muse
cd muse

# Install dependencies
go mod download

# Build binary
go build -o muse cmd/muse/main.go

# Run tests
go test ./...

# Install locally for development
go install ./cmd/muse
```

### Development Workflow

1. **Local Testing**: Use `go run cmd/muse/main.go` for quick iteration
2. **Hook Testing**: Install in test repository for Git integration testing
3. **Configuration Testing**: Test various config sources and precedence
4. **Cross-Platform**: Test on macOS, Linux, and Windows

### Build Configuration

- **GoReleaser**: Automated builds for multiple platforms
- **Binary Distribution**: Single-file executables for easy installation
- **Version Embedding**: Build-time version, commit, and date injection

## Technical Constraints

### Platform Constraints

- **Cross-Platform**: Must work on macOS, Linux, and Windows
- **Git Dependency**: Requires Git 2.0+ for hook support
- **File System**: Must handle different path separators and permissions

### Performance Constraints

- **Hook Speed**: Git hook execution must be fast (< 2 seconds typically)
- **Network Timeouts**: LLM API calls need appropriate timeout handling
- **Memory Usage**: Minimal memory footprint for Git integration

### Security Constraints

- **API Key Management**: Secure handling of LLM provider credentials
- **Input Sanitization**: Safe processing of Git diffs before LLM submission
- **Output Validation**: Validation of LLM responses before Git integration

### API Constraints

- **Rate Limiting**: Must handle LLM provider rate limits gracefully
- **Token Limits**: Respect model context window limitations
- **Network Dependencies**: Graceful degradation when offline

## Dependencies

### Production Dependencies

```go
require (
    github.com/briandowns/spinner v1.23.1       // CLI spinners
    github.com/invopop/jsonschema v0.12.0       // JSON schema generation
    github.com/knadh/koanf v1.5.0               // Configuration management
    github.com/openai/openai-go v0.1.0-alpha.25 // OpenAI API client
    github.com/urfave/cli/v2 v2.27.4            // CLI framework
)
```

### Development Dependencies

- **github.com/stretchr/testify**: Testing framework and assertions
- **golang.org/x/sys**: System-specific functionality
- **golang.org/x/term**: Terminal handling

### Indirect Dependencies

- **github.com/fatih/color**: Terminal color support
- **github.com/mitchellh/mapstructure**: Configuration mapping
- **gopkg.in/yaml.v3**: YAML processing
- Various JSON and utility libraries

## Tool Usage Patterns

### Configuration Management Pattern

```go
// Hierarchical configuration loading
k := koanf.New(".")

// 1. Load config file (if exists)
k.Load(file.Provider(configPath), yaml.Parser())

// 2. Override with environment variables
k.Load(env.Provider("MUSE_", ".", transformFunc), nil)

// 3. Unmarshal to struct
k.Unmarshal("", &config)
```

### LLM Provider Pattern

```go
// Register providers at init time
func init() {
    llm.RegisterProvider("openai", &OpenAIProvider{})
    llm.RegisterProvider("anthropic", &AnthropicProvider{})
}

// Create service from configuration
service, err := llm.NewLLMService(&config.LLM)
```

### Template Processing Pattern

```go
// Compile template with dynamic data
template, err := tm.CompileTemplate(style)

// Execute with diff data
var prompt bytes.Buffer
template.Template.Execute(&prompt, map[string]interface{}{
    "Diff": gitDiff,
    "Style": commitStyle,
})
```

### Error Handling Pattern

- **Graceful Degradation**: Continue with manual commit if AI fails
- **Contextual Errors**: Provide helpful error messages with context
- **Logging**: Structured logging with configurable levels
- **Retry Logic**: Implement retries for transient failures

### Git Hook Integration Pattern

```bash
#!/bin/sh
# prepare-commit-msg hook
exec muse prepare-commit-msg "$@"
```

### Testing Patterns

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test Git hook integration end-to-end
- **Configuration Tests**: Test various config scenarios
- **Provider Tests**: Mock LLM provider responses

### Distribution Pattern

- **Single Binary**: No external dependencies for end users
- **Cross-Platform**: Native binaries for major platforms
- **Package Managers**: Support for Homebrew, apt, etc.
- **GitHub Releases**: Automated release process with GoReleaser

### Security Patterns

- **Environment Variables**: Prefer env vars for sensitive data
- **Config File Permissions**: Validate file permissions for security
- **Input Validation**: Sanitize all external inputs
- **Output Escaping**: Escape LLM outputs before Git integration
