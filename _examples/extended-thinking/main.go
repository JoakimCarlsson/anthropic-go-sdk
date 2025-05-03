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

	// Create a complex math query
	userQuery := "Calculate the value of (342 * 15) - (27^2) + sqrt(169) + the sum of the first 10 prime numbers."

	// Create conversation history
	messages := []models.MessageParam{
		models.NewUserMessage(models.CreateTextBlock(userQuery)),
	}

	// Create message request with extended thinking enabled
	req := models.MessageRequest{
		Model:     models.Claude37Sonnet,
		MaxTokens: 4096,
		Messages:  messages,
		Thinking:  models.EnableThinking(3096),
		Stream:    true,
	}

	// Create streaming request
	ctx := context.Background()
	stream, err := client.CreateMessageStream(ctx, req)
	if err != nil {
		fmt.Println("\n[ERROR]")
		fmt.Printf("Failed to create message stream: %v\n", err)

		apiErr, ok := err.(*anthropic.APIError)
		if ok {
			fmt.Println("\n[ERROR DETAILS]")
			fmt.Printf("Status Code: %d\n", apiErr.StatusCode)
			fmt.Printf("Error Type: %s\n", apiErr.Type)
			if apiErr.Code != "" {
				fmt.Printf("Error Code: %s\n", apiErr.Code)
			}
			if apiErr.Message != "" {
				fmt.Printf("Error Message: %s\n", apiErr.Message)
			}
			if apiErr.Param != "" {
				fmt.Printf("Invalid Parameter: %s\n", apiErr.Param)
			}
		}

		os.Exit(1)
	}

	fmt.Println("\n[Query]:", userQuery)
	fmt.Println("\n--- Claude's thinking process ---")

	for stream.Next() {
		event := stream.Current()

		switch event.Type {
		case streaming.ContentBlockStartEvent:
			if event.ContentBlock != nil {
				if event.ContentBlock.ThinkingContent != nil {
					fmt.Println("\n[Thinking]:")
				} else if event.ContentBlock.RedactedThinkingContent != nil {
					fmt.Println("\n[Redacted Thinking] (not displayed)")
				} else if event.ContentBlock.TextContent != nil {
					fmt.Println("\n[Answer]:")
				}
			}

		case streaming.ContentBlockDeltaEvent:
			if event.Delta != nil {
				if event.Delta.Text != "" {
					fmt.Print(event.Delta.Text)
				} else if event.Delta.Thinking != "" {
					fmt.Print(event.Delta.Thinking)
				}
			}

		case streaming.ContentBlockStopEvent:
			fmt.Println()

		case streaming.MessageStopEvent:
			message := stream.Message()
			fmt.Printf("\n[Message complete. Stop reason: %s]\n", message.StopReason)
			if message.Usage.InputTokens > 0 || message.Usage.OutputTokens > 0 {
				fmt.Printf("[Tokens: %d input, %d output]\n",
					message.Usage.InputTokens, message.Usage.OutputTokens)
			}
		}
	}

	if err := stream.Err(); err != nil {
		fmt.Printf("Stream error: %v\n", err)
		os.Exit(1)
	}

	message := stream.Message()
	fmt.Printf("\n--- Thinking and response complete ---\n")
	fmt.Printf("Total tokens: %d input, %d output\n",
		message.Usage.InputTokens, message.Usage.OutputTokens)
}
