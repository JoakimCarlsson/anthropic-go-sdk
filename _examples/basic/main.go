package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
)

func main() {
	// Create client
	client := anthropic.NewClient()

	// Create message request
	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 1024,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock("What is a quaternion?")),
		},
	}

	// Call API
	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print response
	for _, block := range resp.Content {
		if block.TextContent != nil {
			fmt.Println(block.TextContent.Text)
		}
	}

	fmt.Printf("\nUsage: %d input tokens, %d output tokens\n",
		resp.Usage.InputTokens, resp.Usage.OutputTokens)
}
