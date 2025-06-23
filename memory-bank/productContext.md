# productContext.md

**Purpose:**
Describes why this project exists, the problems it solves, how it should work, and user experience goals.

---

## Problem Statement

### Developer Pain Points

1. **Inconsistent Commit Messages**: Developers often write vague, inconsistent, or poorly formatted commit messages
2. **Mental Overhead**: Writing meaningful commit messages requires mental effort that interrupts the coding flow
3. **Team Standards**: Enforcing commit message conventions across teams is difficult without automation
4. **Time Consumption**: Crafting good commit messages takes time away from actual development work
5. **Knowledge Gap**: Not all developers know best practices for commit message formatting

### Technical Challenges

- Manual commit message writing leads to poor Git history readability
- Inconsistent formats make automated tooling and release note generation difficult
- Code review processes are hindered by unclear commit descriptions
- Project maintenance becomes harder with poor commit documentation

## Rationale

### Why Muse is Needed

1. **Automation**: Leverages AI to automatically generate high-quality commit messages
2. **Standardization**: Enforces consistent commit message formats across teams and projects
3. **Intelligence**: Analyzes actual code changes to produce contextually relevant messages
4. **Flexibility**: Supports multiple commit styles to match team preferences
5. **Integration**: Seamlessly fits into existing Git workflows without disruption

### Business Value

- Improved code maintainability through better Git history
- Reduced onboarding time for new developers
- Enhanced code review processes
- Better release documentation and changelogs
- Increased developer productivity and satisfaction

## User Experience Goals

### Primary UX Objectives

1. **Seamless Integration**: Should feel like a natural part of the Git workflow
2. **User Control**: Always allow developers to review and modify generated messages
3. **Quick Setup**: Simple installation and configuration process
4. **Immediate Value**: Generate useful commit messages from the first use
5. **Predictable Behavior**: Consistent, reliable output that developers can trust

### Target User Personas

- **Individual Developers**: Want better commit messages without extra effort
- **Team Leads**: Need to enforce commit standards across their teams
- **Open Source Maintainers**: Require consistent, professional commit history
- **Enterprise Teams**: Need compliance with corporate development standards

## How It Should Work

### Core Workflow

1. **Developer makes changes** and stages them with `git add`
2. **Developer runs `git commit`** (without message)
3. **Muse analyzes the staged diff** using the prepare-commit-msg hook
4. **AI generates appropriate commit message** based on changes and configured style
5. **Developer reviews the generated message** in their editor
6. **Developer can accept, modify, or reject** the message before committing

### Configuration Experience

1. **One-time setup**: Run `muse configure` to create initial configuration
2. **Provider setup**: Configure API keys for chosen LLM provider
3. **Style selection**: Choose preferred commit message style (conventional, gitmoji, etc.)
4. **Hook installation**: Run `muse install` to set up Git hooks

### Command-Line Interface

- `muse install`: Install Git hooks in current repository
- `muse uninstall`: Remove Git hooks from current repository
- `muse configure`: Create or update configuration file
- `muse status`: Show current installation and configuration status
- `muse version`: Display version information

### Preview and Control

- **Preview mode**: See generated message before it's applied
- **Dry run mode**: Test configuration without actually committing
- **Manual override**: Always allow manual editing of generated messages
- **Fallback behavior**: Graceful degradation when AI service is unavailable

### Style Support

- **Conventional Commits**: feat(api): add user authentication endpoint
- **Gitmoji**: ✨ feat(api): add user authentication endpoint
- **Default**: Simple, clear format without strict conventions
- **Custom**: Extensible template system for organization-specific formats
