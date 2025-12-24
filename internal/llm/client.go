package llm

import "context"

type Client interface {
	Check(ctx context.Context) error
	Generate(ctx context.Context, prompt string) (string, error)
	Embed(ctx context.Context, text string) ([]float32, error)
}

type Config struct {
	BaseURL    string
	Model      string
	EmbedModel string
}
