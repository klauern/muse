# Muse: AI-Powered Git Commit Message Generator

Muse is an intelligent tool designed to automatically generate meaningful and consistent Git commit messages using AI. It integrates with your Git workflow to analyze your changes and produce appropriate commit messages based on the diff.

## How It Works

1. Muse hooks into your Git commit process.
2. When you stage changes and attempt to commit, Muse analyzes the diff.
3. It uses a Language Model (LLM) to generate a commit message based on the changes.
4. The generated message is presented for your review and can be used as-is or modified.

## Features

- Supports multiple LLM providers (OpenAI, Anthropic, Ollama)
- Configurable commit message styles (Conventional, Gitmojis, Default)
- Retrieval-Augmented Generation (RAG) for context-aware commit messages
- Customizable configuration

## Installation

[Add installation instructions here]

## Configuration

Muse can be configured using a YAML file. The configuration file is typically located at `$HOME/.config/muse/config.yaml` or can be specified using the `--config` flag.

### Example Configuration

```yaml
hook:
  type: default
  commit_style: conventional
  dry_run: false
  preview: true

llm:
  provider: anthropic
  config:
    api_key: your_api_key_here
    model: claude-2
```

### Configuration Options

- `hook.type`: The type of hook to use (default, llm)
- `hook.commit_style`: The style of commit messages to generate (conventional, gitmojis, default)
- `hook.dry_run`: Run without actually committing
- `hook.preview`: Preview the generated commit message before applying
- `llm.provider`: The LLM provider to use (anthropic, openai, ollama)
- `llm.config`: Provider-specific configuration options

## Usage

Once configured, Muse will automatically generate commit messages when you run `git commit`. You can also use the Muse CLI for more control:

```
muse generate --provider anthropic --style conventional
```

For more information on available commands and options, run:

```
muse --help
```

## Contributing

[Add contribution guidelines here]

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
