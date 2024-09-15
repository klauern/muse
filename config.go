package pre_commit_llm

type ModelConfig struct {
	ModelName         string
	Temperature       float32
	MaxTokens         int
	TopP              float32
	FrequencyPenalty  float32
	PresencePenalty   float32
	StopSequences     []string
}

type LLMConfig struct {
	Provider string
	Config   map[string]interface{}
}

