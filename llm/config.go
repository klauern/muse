package llm

// LLMConfig is a generic configuration structure for LLM providers
type LLMConfig[T any] struct {
	Provider string
	Config   T
}

// OpenAIConfig holds configuration specific to OpenAI
type OpenAIConfig struct {
	APIKey string
	Model  string
	// Add other OpenAI-specific fields as needed
}

// AnthropicConfig holds configuration specific to Anthropic
type AnthropicConfig struct {
	APIKey string
	Model  string
	// Add other Anthropic-specific fields as needed
}

// OllamaConfig holds configuration specific to Ollama
type OllamaConfig struct {
	Model string
	// Add other Ollama-specific fields as needed
}
