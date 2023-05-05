package clients

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func AskChatGPT(ctx context.Context, token, quesiton string) (string, error) {
	client := openai.NewClient(token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: quesiton,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("chatgpt error: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
