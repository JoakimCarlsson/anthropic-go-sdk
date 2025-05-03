package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
	"github.com/joakimcarlsson/anthropic-sdk/streaming"
)

func main() {
	// Create client
	client := anthropic.NewClient()

	// Define search tool
	searchTool := models.NewTool(
		"search",
		"Search for information on a topic",
		models.SimpleJSONSchema(
			map[string]models.Property{
				"query": models.NewProperty("string", "The search query"),
			},
			[]string{"query"},
		),
	)

	// Create conversation history
	messages := []models.MessageParam{
		models.NewUserMessage(models.CreateTextBlock("Who won the world series in 2023? And give me a brief summary of the series.")),
	}

	// Create message request
	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     []models.Tool{searchTool},
		Stream:    true,
	}

	// Loop to handle tool use
	ctx := context.Background()
	for {
		// Create streaming request
		stream, err := client.CreateMessageStream(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n[Assistant]:")

		// Process streaming events
		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case streaming.MessageStartEvent:
				fmt.Println("Starting new message...")

			case streaming.ContentBlockStartEvent:
				if event.ContentBlock != nil && event.ContentBlock.ToolUseContent != nil {
					fmt.Printf("[Using tool: %s]\n", event.ContentBlock.ToolUseContent.Name)
				}

			case streaming.ContentBlockDeltaEvent:
				if event.Delta != nil {
					if event.Delta.Text != "" {
						// Print text delta
						fmt.Print(event.Delta.Text)
					} else if event.Delta.PartialJSON != "" {
						// Print tool input JSON as it comes in
						fmt.Print(event.Delta.PartialJSON)
					}
				}

			case streaming.ContentBlockStopEvent:
				fmt.Println() // Add newline when content block is complete

			case streaming.MessageStopEvent:
				// Get final message with accumulated data
				fmt.Printf("\n[Message complete. Stop reason: %s]\n", stream.Message().StopReason)
				if stream.Message().Usage.InputTokens > 0 || stream.Message().Usage.OutputTokens > 0 {
					fmt.Printf("[Tokens: %d input, %d output]\n",
						stream.Message().Usage.InputTokens, stream.Message().Usage.OutputTokens)
				}
			}
		}

		// Check for errors
		if err := stream.Err(); err != nil {
			fmt.Printf("Stream error: %v\n", err)
			os.Exit(1)
		}

		// Get the final message with all accumulated data
		message := stream.Message()

		// Process tool uses from the message
		toolResults := []models.ContentBlock{}
		for _, block := range message.Content {
			if block.ToolUseContent != nil {
				toolUseID := block.ToolUseContent.ID
				toolName := block.ToolUseContent.Name
				toolInput := block.ToolUseContent.Input

				// Process tool use
				var result string
				switch toolName {
				case "search":
					result = handleSearchTool(toolInput)
				default:
					result = fmt.Sprintf("Unknown tool: %s", toolName)
				}

				fmt.Printf("[Tool result: %s]\n", result)

				// Create tool result
				toolResults = append(
					toolResults,
					models.CreateToolResultBlock(toolUseID, result, false),
				)
			}
		}

		// If no tools were used, we're done
		if len(toolResults) == 0 {
			break
		}

		// Add the response and tool results to messages
		messages = append(messages, models.MessageParam{
			Role:    models.AssistantRole,
			Content: message.Content,
		})
		messages = append(messages, models.MessageParam{
			Role:    models.UserRole,
			Content: toolResults,
		})

		// Update request for next iteration
		req.Messages = messages
	}
}

func handleSearchTool(inputData interface{}) string {
	// Check if inputData is nil or empty
	if inputData == nil {
		return "Error: Received empty tool input"
	}

	// Parse input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return fmt.Sprintf("Error marshaling input data: %v", err)
	}

	// Check for empty JSON object
	if string(inputBytes) == "{}" || string(inputBytes) == "null" {
		return "Error: Received empty JSON input for search tool"
	}

	var input struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	// Check if query is empty
	if input.Query == "" {
		return "Error: No search query provided"
	}

	// Mock search results based on query
	query := strings.ToLower(input.Query)

	if strings.Contains(query, "world series") && strings.Contains(query, "2023") {
		return `The Texas Rangers won the 2023 World Series, defeating the Arizona Diamondbacks 4 games to 1.
This was the Rangers' first World Series championship in franchise history.
Game 1: Rangers 6, Diamondbacks 5 (11 innings)
Game 2: Rangers 9, Diamondbacks 1
Game 3: Diamondbacks 3, Rangers 1
Game 4: Rangers 11, Diamondbacks 7
Game 5: Rangers 5, Diamondbacks 0
Rangers shortstop Corey Seager was named World Series MVP.`
	}

	return "No relevant information found for the query: " + input.Query
}
