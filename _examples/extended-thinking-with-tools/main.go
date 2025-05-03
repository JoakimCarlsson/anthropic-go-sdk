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

	// Define calculator tool
	calculatorTool := models.NewTool(
		"calculate",
		"A simple calculator for basic arithmetic operations",
		models.SimpleJSONSchema(
			map[string]models.Property{
				"operation": models.NewEnumProperty(
					"The arithmetic operation to perform",
					[]string{"add", "subtract", "multiply", "divide"},
				),
				"operands": models.NewProperty("array", "The operands for the operation"),
			},
			[]string{"operation", "operands"},
		),
	)

	// Create a complex request that requires thinking and tool use
	userQuery := "I need to calculate the result of multiplying 342 by 15, then subtracting 729, then dividing by 3. Can you help me with that?"

	// Create conversation history
	messages := []models.MessageParam{
		models.NewUserMessage(models.CreateTextBlock(userQuery)),
	}

	// Create message request with extended thinking enabled
	req := models.MessageRequest{
		Model:     models.Claude37Sonnet,
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     []models.Tool{calculatorTool},
		Thinking:  models.EnableThinking(3096),
		Stream:    true,
	}

	// Loop to handle tool use
	ctx := context.Background()
	fmt.Println("\n[User]:", userQuery)

	// Variable to store the final stream for usage statistics
	var finalStream *streaming.MessageStream

	for {
		// Create streaming request
		stream, err := client.CreateMessageStream(ctx, req)
		if err != nil {
			fmt.Println("\n[ERROR]")
			fmt.Printf("Failed to create message stream: %v\n", err)

			// Try to provide more helpful troubleshooting information
			apiErr, ok := err.(*anthropic.APIError)
			if ok {
				// Print the full error details including raw response body
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

		// Save the current stream for later usage reporting
		finalStream = stream

		fmt.Println("\n[Assistant]:")

		// Process streaming events
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
						fmt.Println("\n[Text]:")
					} else if event.ContentBlock.ToolUseContent != nil {
						fmt.Printf("\n[Using tool: %s]\n", event.ContentBlock.ToolUseContent.Name)
					}
				}

			case streaming.ContentBlockDeltaEvent:
				if event.Delta != nil {
					if event.Delta.Text != "" {
						// Print text delta
						fmt.Print(event.Delta.Text)
					} else if event.Delta.Thinking != "" {
						// Print thinking delta
						fmt.Print(event.Delta.Thinking)
					} else if event.Delta.PartialJSON != "" {
						// Print tool input JSON as it comes in
						fmt.Print(event.Delta.PartialJSON)
					}
					// Skip signature delta
				}

			case streaming.ContentBlockStopEvent:
				fmt.Println() // Add newline when content block is complete

			case streaming.MessageStopEvent:
				// Message is complete
				fmt.Printf("\n[Message complete. Stop reason: %s]\n", stream.Message().StopReason)
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
				if toolName == "calculate" {
					result = handleCalculatorTool(toolInput)
				} else {
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

		// Save the assistant message as is, including thinking blocks
		messages = append(messages, models.MessageParam{
			Role:    models.AssistantRole,
			Content: message.Content,
		})

		// Create a user message with ONLY the tool results - DO NOT include thinking blocks
		messages = append(messages, models.MessageParam{
			Role:    models.UserRole,
			Content: toolResults,
		})

		// Update request for next iteration
		req.Messages = messages
	}

	// Display final usage statistics
	if finalStream != nil {
		finalMessage := finalStream.Message()
		fmt.Printf("\n--- Thinking and response complete ---\n")
		fmt.Printf("Total tokens: %d input, %d output\n",
			finalMessage.Usage.InputTokens, finalMessage.Usage.OutputTokens)
	}
}

func handleCalculatorTool(inputData interface{}) string {
	// Parse input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return fmt.Sprintf("Error marshaling input data: %v", err)
	}

	var input struct {
		Operation string    `json:"operation"`
		Operands  []float64 `json:"operands"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	// Validate input
	if input.Operation == "" {
		return "Error: No operation specified"
	}
	if len(input.Operands) < 2 {
		return "Error: Need at least two operands"
	}

	// Perform calculation
	var result float64
	switch strings.ToLower(input.Operation) {
	case "add":
		result = input.Operands[0]
		for _, op := range input.Operands[1:] {
			result += op
		}
	case "subtract":
		result = input.Operands[0]
		for _, op := range input.Operands[1:] {
			result -= op
		}
	case "multiply":
		result = input.Operands[0]
		for _, op := range input.Operands[1:] {
			result *= op
		}
	case "divide":
		result = input.Operands[0]
		for _, op := range input.Operands[1:] {
			if op == 0 {
				return "Error: Division by zero"
			}
			result /= op
		}
	default:
		return fmt.Sprintf("Error: Unknown operation %s", input.Operation)
	}

	return fmt.Sprintf("Result: %.2f", result)
}
