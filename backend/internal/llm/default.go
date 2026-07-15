package llm

import "context"

type defaultClient struct{}

func (defaultClient) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	return ChatCompletion(ctx, messages)
}

func (defaultClient) StreamChat(ctx context.Context, messages []Message) (<-chan string, <-chan error) {
	return StreamChat(ctx, messages)
}

var DefaultClient Client = defaultClient{}
