package pre_commit_llm

// ModelConfig represents common configuration options for language models
type ModelConfig struct {
	ModelName         string
	Temperature       float32
	MaxTokens         int
	TopP              float32
	FrequencyPenalty  float32
	PresencePenalty   float32
	StopSequences     []string
}

