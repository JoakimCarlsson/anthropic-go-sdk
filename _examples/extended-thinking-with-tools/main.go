package main

import (
	"bufio"
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
	client := anthropic.NewClient()

	weatherSchema := models.SimpleJSONSchema(
		map[string]models.Property{
			"location": models.NewProperty("string", "The city and state, e.g. San Francisco, CA"),
		},
		[]string{"location"},
	)

	calculatorSchema := models.SimpleJSONSchema(
		map[string]models.Property{
			"expression": models.NewProperty("string", "The mathematical expression to evaluate"),
		},
		[]string{"expression"},
	)

	tools := []models.Tool{
		{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			InputSchema: weatherSchema,
		},
		{
			Name:        "calculate",
			Description: "Perform a mathematical calculation",
			InputSchema: calculatorSchema,
		},
	}

	var messages []models.MessageParam

	ctx := context.Background()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Interactive Chat with Claude (with Extended Thinking and Tools) ===")
	fmt.Println("Type 'exit' to end the conversation.")
	fmt.Println()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()

		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Exiting chat...")
			break
		}

		userMessage := models.MessageParam{
			Role: models.UserRole,
			Content: []models.ContentBlock{
				models.CreateTextBlock(userInput),
			},
		}

		messages = append(messages, userMessage)

		req := models.MessageRequest{
			Model:     models.Claude37Sonnet,
			MaxTokens: 4000,
			Messages:  messages,
			Tools:     tools,
			Thinking:  models.EnableThinking(2000),
			Stream:    true,
		}

		fmt.Println("\nClaude is thinking...")

		stream, err := client.CreateMessageStream(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("\nClaude:")

		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case streaming.MessageStartEvent:

			case streaming.ContentBlockStartEvent:
				if event.ContentBlock != nil {
					if event.ContentBlock.ThinkingContent != nil {
						fmt.Print("\n[Thinking] ")
					} else if event.ContentBlock.RedactedThinkingContent != nil {
						fmt.Print("\n[Redacted Thinking] ")
					} else if event.ContentBlock.TextContent != nil {
					} else if event.ContentBlock.ToolUseContent != nil {
						fmt.Printf("\n[Using tool: %s] ", event.ContentBlock.ToolUseContent.Name)
					}
				}

			case streaming.ContentBlockDeltaEvent:
				if event.Delta != nil {
					if event.Delta.Text != "" {
						fmt.Print(event.Delta.Text)
					} else if event.Delta.Thinking != "" {
						fmt.Print(event.Delta.Thinking)
					} else if event.Delta.PartialJSON != "" {
						fmt.Print(event.Delta.PartialJSON)
					}
				}

			case streaming.ContentBlockStopEvent:
				if event.ContentBlock != nil && event.ContentBlock.ThinkingContent != nil {
					fmt.Println("\n")
				}

			case streaming.MessageStopEvent:
			}
		}

		if err := stream.Err(); err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		message := stream.Message()

		claudeMessage := models.MessageParam{
			Role:    models.AssistantRole,
			Content: message.Content,
		}
		messages = append(messages, claudeMessage)

		var toolResults []models.ContentBlock
		for _, block := range message.Content {
			if block.ToolUseContent != nil {
				toolCall := block.ToolUseContent
				switch toolCall.Name {
				case "calculate":
					input, _ := json.Marshal(toolCall.Input)
					var parsedInput map[string]interface{}
					json.Unmarshal(input, &parsedInput)

					expression, _ := parsedInput["expression"].(string)
					fmt.Printf("\n\nCalculating: %s\n", expression)

					var result string
					if expression == "234 * 78" {
						result = "18252"
					} else {
						result = fmt.Sprintf("Result: %s (simplified calculation)", expression)
					}

					fmt.Printf("Result: %s\n", result)
					toolResults = append(toolResults, models.CreateToolResultBlock(
						toolCall.ID,
						result,
						false,
					))

				case "get_weather":
					input, _ := json.Marshal(toolCall.Input)
					var parsedInput map[string]interface{}
					json.Unmarshal(input, &parsedInput)

					location, _ := parsedInput["location"].(string)
					fmt.Printf("\n\nGetting weather for: %s\n", location)

					weatherData := map[string]interface{}{
						"temperature": 68,
						"condition":   "Partly Cloudy",
						"humidity":    72,
						"wind":        "10 mph",
					}

					weatherJSON, _ := json.Marshal(weatherData)
					fmt.Printf("Weather: %s\n", string(weatherJSON))

					toolResults = append(toolResults, models.CreateToolResultBlock(
						toolCall.ID,
						string(weatherJSON),
						false,
					))
				}
			}
		}

		if len(toolResults) > 0 {
			toolResultMessage := models.MessageParam{
				Role:    models.UserRole,
				Content: toolResults,
			}
			messages = append(messages, toolResultMessage)

			req = models.MessageRequest{
				Model:     models.Claude37Sonnet,
				MaxTokens: 4000,
				Messages:  messages,
				Tools:     tools,
				Thinking:  models.EnableThinking(2000),
				Stream:    true,
			}

			fmt.Println("\nSending tool results to Claude...")

			stream, err := client.CreateMessageStream(ctx, req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			fmt.Println("\nClaude:")

			for stream.Next() {
				event := stream.Current()

				switch event.Type {
				case streaming.ContentBlockStartEvent:
					if event.ContentBlock != nil {
						if event.ContentBlock.ThinkingContent != nil {
							fmt.Print("\n[Thinking] ")
						} else if event.ContentBlock.RedactedThinkingContent != nil {
							fmt.Print("\n[Redacted Thinking] ")
						}
					}

				case streaming.ContentBlockDeltaEvent:
					if event.Delta != nil {
						if event.Delta.Text != "" {
							fmt.Print(event.Delta.Text)
						} else if event.Delta.Thinking != "" {
							fmt.Print(event.Delta.Thinking)
						} else if event.Delta.PartialJSON != "" {
							fmt.Print(event.Delta.PartialJSON)
						}
					}

				case streaming.ContentBlockStopEvent:
					if event.ContentBlock != nil && event.ContentBlock.ThinkingContent != nil {
						fmt.Println("\n")
					}
				}
			}

			if err := stream.Err(); err != nil {
				fmt.Printf("\nError: %v\n", err)
				continue
			}

			finalMessage := stream.Message()

			finalClaudeMessage := models.MessageParam{
				Role:    models.AssistantRole,
				Content: finalMessage.Content,
			}
			messages = append(messages, finalClaudeMessage)
		}

		fmt.Println("\n")
	}
}
