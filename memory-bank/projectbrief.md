# projectbrief.md

**Purpose:**
Foundation document that shapes all other files.
Defines core requirements and goals.
Source of truth for project scope.

---

## Project Name

**Muse: AI-Powered Git Commit Message Generator**

## Overview

Muse is an intelligent CLI tool that automatically generates meaningful and consistent Git commit messages using AI. It integrates seamlessly with Git workflows by analyzing staged changes and producing appropriate commit messages based on the diff, supporting multiple LLM providers and customizable commit message styles.

## Core Requirements

1. **Git Integration**: Hook into Git commit process to analyze diffs automatically
2. **AI-Powered Generation**: Use LLM providers (OpenAI, Anthropic, Ollama) to generate commit messages
3. **Multiple Commit Styles**: Support conventional commits, gitmoji, and default formats
4. **Configuration Management**: YAML-based configuration with environment variable support
5. **Preview and Control**: Allow users to preview and approve generated messages
6. **CLI Interface**: Provide comprehensive command-line interface for all operations
7. **Cross-Platform**: Work on macOS, Linux, and Windows environments

## Goals

1. **Improve Commit Quality**: Generate consistent, meaningful commit messages that follow best practices
2. **Reduce Developer Friction**: Eliminate the mental overhead of writing commit messages
3. **Enforce Standards**: Help teams maintain consistent commit message conventions
4. **Flexibility**: Support different workflows and commit message styles
5. **User Control**: Always allow user review and modification of generated messages

## Scope

### In Scope

- Git hook integration (prepare-commit-msg)
- AI-powered commit message generation
- Multiple LLM provider support
- Configurable commit message styles (conventional, gitmoji, default)
- YAML configuration with environment variable overrides
- CLI commands for installation, configuration, and management
- Preview and dry-run capabilities
- JSON schema validation for structured outputs

### Out of Scope

- Direct Git repository management
- Commit message editing beyond generation
- Integration with Git GUIs (initial version)
- Real-time collaboration features
- Commit message history analysis
- Integration with issue tracking systems (initial version)
