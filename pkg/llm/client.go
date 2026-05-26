package llm

import (
	"context"
)

type Client interface {
	StreamCompletion(ctx context.Context, systemPrompt, userPrompt string, callback func(string)) error
}
