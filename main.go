package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	openai "github.com/sashabaranov/go-openai"
)

var messages = []openai.ChatCompletionMessage{}

func chat(c *openai.Client, ctx context.Context, msg string) {
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	}

	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	defer stream.Close()

	first := true
	re := regexp.MustCompile(`[\s|\n]+`)
	resContext := ""

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: resContext,
			})
			fmt.Println()
			return
		}
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		txt := response.Choices[0].Delta.Content
		if first && (txt == "" || re.MatchString(txt)) {
			continue
		}
		first = false
		fmt.Print(txt)
		resContext += txt
	}

}

func main() {
	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	for {
		fmt.Print("Q: ")
		var msg string
		fmt.Scanln(&msg)
		chat(c, ctx, msg)
	}
}
