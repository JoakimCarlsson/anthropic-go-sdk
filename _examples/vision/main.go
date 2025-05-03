package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
)

func main() {
	// Create client
	client := anthropic.NewClient()

	imagePath := "\\rickastley\\nevergonnagiveyouup.png"

	// Encode image
	imageData, mediaType, err := models.Base64EncodeImage(imagePath)
	if err != nil {
		fmt.Printf("Error encoding image: %v\n", err)
		os.Exit(1)
	}

	// Create image source
	imageSource := models.NewBase64ImageSource(mediaType, imageData)

	// Create multimodal message
	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 1024,
		Messages: []models.MessageParam{
			models.NewUserMessage(
				models.CreateImageBlock(imageSource),
				models.CreateTextBlock("What's in this image? Please describe it in detail."),
			),
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
	fmt.Println("[Image: " + filepath.Base(imagePath) + "]")
	fmt.Println("\n[Assistant]:")
	for _, block := range resp.Content {
		if block.TextContent != nil {
			fmt.Println(block.TextContent.Text)
		}
	}
}
