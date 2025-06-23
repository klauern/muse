# systemPatterns.md

**Purpose:**
Documents system architecture, key technical decisions, design patterns, component relationships, and critical implementation paths.

---

## System Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Git Process   │    │   Muse CLI      │    │   LLM Provider  │
│                 │    │                 │    │                 │
│ git commit ──────────▶ Hook Handler ──────▶ AI Service      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        │                       ▼                       │
        │              ┌─────────────────┐               │
        │              │ Template Engine │               │
        │              └─────────────────┘               │
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Commit Message  │◀───│ Config Manager  │    │ Generated JSON  │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components

1. **CLI Application** (`cmd/muse/main.go`)
   - Entry point using urfave/cli framework
   - Command routing and global configuration loading
   - Version information and help system

2. **Hook System** (`hooks/`)
   - Git hook installation and management
   - prepare-commit-msg hook implementation
   - Repository-specific hook handling

3. **Configuration Management** (`config/`)
   - YAML-based configuration with environment variable support
   - XDG Base Directory specification compliance
   - Provider-specific configuration handling

4. **LLM Service Layer** (`llm/`)
   - Provider-agnostic interface design
   - Plugin-style provider registration
   - OpenAI provider implementation (extensible to others)

5. **Template System** (`templates/`)
   - Multiple commit style support (conventional, gitmoji, default)
   - JSON schema generation for structured outputs
   - Dynamic template compilation and execution

## Key Technical Decisions

### Language and Framework Choices

- **Go**: Chosen for cross-platform compatibility, single binary distribution, and strong CLI tooling ecosystem
- **urfave/cli**: Provides robust command-line interface with subcommands and flag parsing
- **koanf**: Configuration management with multiple source support (files, environment, etc.)

### Architecture Patterns

- **Provider Pattern**: LLM services implement a common interface for extensibility
- **Template Method**: Commit message generation follows a structured template approach
- **Plugin Registration**: LLM providers register themselves for dynamic loading
- **Configuration Hierarchy**: File → Environment → Defaults with clear precedence

### API and Data Flow

- **Structured Outputs**: Uses JSON schema to ensure consistent LLM responses
- **Git Integration**: Leverages Git's prepare-commit-msg hook for seamless workflow integration
- **Stateless Design**: Each commit message generation is independent and stateless

## Design Patterns in Use

### 1. **Provider Pattern**

```go
type LLMService interface {
    GenerateCommitMessage(ctx context.Context, diff string, style templates.CommitStyle) (string, error)
}

type LLMProvider interface {
    NewService(config map[string]interface{}) (LLMService, error)
}
```

### 2. **Registry Pattern**

```go
var providers = make(map[string]LLMProvider)

func RegisterProvider(name string, provider LLMProvider) {
    providers[name] = provider
}
```

### 3. **Template Method Pattern**

- Base template structure for all commit styles
- Configurable parameters for different formats
- Runtime template compilation with dynamic data injection

### 4. **Builder Pattern**

- Configuration building with hierarchical sources
- Template compilation with dynamic content
- CLI application construction with multiple commands

## Component Relationships

### Configuration Flow

```
CLI Flags ──┐
            ├─▶ Config Manager ──▶ Unified Config ──▶ Services
Env Vars ───┤
            │
YAML File ──┘
```

### Message Generation Flow

```
Git Diff ──▶ Template Manager ──▶ LLM Service ──▶ JSON Response ──▶ Formatted Message
     │              │                    │               │
     │              ▼                    │               │
     │    ┌─────────────────┐             │               │
     │    │ Style Selection │             │               │
     │    └─────────────────┘             │               │
     │                                    │               │
     └──▶ Schema Generation ──────────────┘               │
                     │                                   │
                     ▼                                   │
            ┌─────────────────┐                          │
            │ JSON Validation │◀─────────────────────────┘
            └─────────────────┘
```

### Command Relationships

```
muse install ──▶ Hook Installation ──▶ Repository Setup
muse configure ──▶ Config Creation ──▶ User Setup
muse status ──▶ State Inspection ──▶ User Feedback
muse uninstall ──▶ Hook Removal ──▶ Cleanup
```

## Critical Implementation Paths

### 1. **Git Hook Integration**

- **Path**: Git commit → prepare-commit-msg hook → muse execution
- **Critical**: Must handle Git's environment variables and file paths correctly
- **Failure Mode**: Hook failures must not break Git commits

### 2. **LLM API Communication**

- **Path**: Diff analysis → API request → JSON response → message extraction
- **Critical**: Network failures, API limits, and malformed responses must be handled
- **Failure Mode**: Graceful degradation to manual commit message entry

### 3. **Configuration Resolution**

- **Path**: Multiple config sources → merged configuration → service initialization
- **Critical**: Config precedence must be consistent and predictable
- **Failure Mode**: Must provide sensible defaults when configuration is missing

### 4. **Template Processing**

- **Path**: Style selection → template compilation → dynamic data injection → final prompt
- **Critical**: Template syntax errors must be caught at compile time
- **Failure Mode**: Default template fallback when custom templates fail

### 5. **JSON Schema Validation**

- **Path**: LLM response → JSON parsing → schema validation → structured data
- **Critical**: Invalid JSON must be handled without breaking the workflow
- **Failure Mode**: Retry with simpler prompt or manual fallback

### Security Considerations

- **API Key Management**: Environment variables preferred over config files
- **Git Hook Security**: Hooks run in user context with repository permissions
- **Input Validation**: Git diffs are sanitized before sending to LLM providers
- **Output Sanitization**: LLM responses are validated before Git integration
