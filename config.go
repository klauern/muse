package muse

type Config struct {
}

type ModelConfig struct {
	ModelName     string
	Temperature   float32
	MaxTokens     int
	TopP          float32
	FrequencyPenalty float32
	PresencePenalty  float32
	StopSequences    []string
}

