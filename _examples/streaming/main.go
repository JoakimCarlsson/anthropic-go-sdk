package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
	"github.com/joakimcarlsson/anthropic-sdk/streaming"
)

func main() {
	// Create client
	client := anthropic.NewClient()

	// Create message request with streaming enabled
	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 1024,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock("Tell me about quantum computing in 3 sentences.")),
		},
		Stream: true,
	}

	// Create streaming request
	ctx := context.Background()
	stream, err := client.CreateMessageStream(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Process stream
	for stream.Next() {
		event := stream.Current()

		// Process event based on type
		switch event.Type {
		case streaming.ContentBlockDeltaEvent:
			if event.Delta != nil {
				fmt.Print(event.Delta.Text)
			}
		case streaming.MessageStopEvent:
			message := stream.Message()
			fmt.Printf("\n\nMessage complete! Usage: %d input tokens, %d output tokens\n",
				message.Usage.InputTokens, message.Usage.OutputTokens)
		}
	}

	// Check for errors
	if err := stream.Err(); err != nil {
		fmt.Printf("Stream error: %v\n", err)
		os.Exit(1)
	}
}
