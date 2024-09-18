package config

type OpenAIConfig struct {
	APIKey  string `env:"OPENAI_API_KEY"`
	Model   string `env:"OPENAI_MODEL"`
	APIBase string `env:"OPENAI_API_BASE" envDefault:"https://api.openai.com/v1"`
}

type AnthropicConfig struct {
	APIKey string `env:"ANTHROPIC_API_KEY"`
	Model  string `env:"ANTHROPIC_MODEL" envDefault:"claude-3-5-sonnet-20240620"`
}

type OllamaConfig struct {
	Model  string `env:"OLLAMA_MODEL"`
	APIUrl string `env:"OLLAMA_API_BASE" envDefault:"http://localhost:11434"`
}
