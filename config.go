package pre_commit_llm

type Config struct {
	HookConfig HookConfig
}

type HookConfig struct {
	Enabled bool
	Type    string
}

type ModelConfig struct {
	ModelName         string
	Temperature       float32
	MaxTokens         int
	TopP              float32
	FrequencyPenalty  float32
	PresencePenalty   float32
	StopSequences     []string
}

