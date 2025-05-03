# Claude Go SDK (Unofficial)

An unofficial, powerful and user-friendly Go SDK for Anthropic's Claude AI models, with enhanced support for Claude 3.7 Sonnet's extended thinking capabilities and tool use.

> **Disclaimer**: This is an unofficial SDK and is not created, maintained, or endorsed by Anthropic. Use at your own discretion.

## Features

- Clean, idiomatic Go API for Claude
- Comprehensive support for Claude's Messages API
- Streaming responses with real-time content processing
- Multimodal support for images
- Tool use (function calling) with intuitive interfaces
- Extended thinking functionality with Claude 3.7 Sonnet
- Robust error handling with detailed error information
- Fully typed models and responses

## Installation

```bash
go get github.com/joakimcarlsson/anthropic-sdk
```

Update the import paths in your code:

```go
import "github.com/joakimcarlsson/anthropic-sdk"
import "github.com/joakimcarlsson/anthropic-sdk/models"
import "github.com/joakimcarlsson/anthropic-sdk/streaming"
```

## Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
)

func main() {
	// Create a client using your API key
	client := api.NewClient(api.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))

	// Create a simple message request
	request := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 1000,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock("Hello, Claude!")),
		},
	}

	// Send the request
	ctx := context.Background()
	response, err := client.CreateMessage(ctx, request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	if len(response.Content) > 0 && response.Content[0].TextContent != nil {
		fmt.Println("Claude:", response.Content[0].TextContent.Text)
	}
}
```

## Featured Examples

### Extended Thinking with Streaming

Claude 3.7 Sonnet's extended thinking allows you to see Claude's reasoning process as it thinks through problems:

```go
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
	client := api.NewClient(api.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))

	// Create a complex math query
	userQuery := "Calculate the value of (342 * 15) - (27^2) + sqrt(169) + the sum of the first 10 prime numbers."

	// Create message request with extended thinking enabled
	req := models.MessageRequest{
		Model:     models.Claude37Sonnet,
		MaxTokens: 4096,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock(userQuery)),
		},
		Thinking: models.EnableThinking(3096), // Enable extended thinking with a token budget
		Stream:   true,
	}

	ctx := context.Background()
	stream, err := client.CreateMessageStream(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n[Query]:", userQuery)
	fmt.Println("\n--- Claude's thinking process ---")

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
		}
	}

	if err := stream.Err(); err != nil {
		fmt.Printf("Stream error: %v\n", err)
		os.Exit(1)
	}
}
```

### Tool Use with Streaming

Tool use (function calling) allows Claude to use your defined functions:

```go
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
	client := api.NewClient(api.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))

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

	// Create a request that requires tool use
	userQuery := "I need to calculate the result of multiplying 342 by 15, then subtracting 729, then dividing by 3. Can you help me with that?"

	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 4096,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock(userQuery)),
		},
		Tools:  []models.Tool{calculatorTool},
		Stream: true,
	}

	ctx := context.Background()
	fmt.Println("\n[User]:", userQuery)

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
			case streaming.ContentBlockStartEvent:
				if event.ContentBlock != nil {
					if event.ContentBlock.TextContent != nil {
						fmt.Println("\n[Text]:")
					} else if event.ContentBlock.ToolUseContent != nil {
						fmt.Printf("\n[Using tool: %s]\n", event.ContentBlock.ToolUseContent.Name)
					}
				}

			case streaming.ContentBlockDeltaEvent:
				if event.Delta != nil {
					if event.Delta.Text != "" {
						fmt.Print(event.Delta.Text)
					} else if event.Delta.PartialJSON != "" {
						fmt.Print(event.Delta.PartialJSON)
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			fmt.Printf("Stream error: %v\n", err)
			os.Exit(1)
		}

		// Get the final message
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
		messages := append(req.Messages, models.MessageParam{
			Role:    models.AssistantRole,
			Content: message.Content,
		})

		// Create a user message with ONLY the tool results
		messages = append(messages, models.MessageParam{
			Role:    models.UserRole,
			Content: toolResults,
		})

		// Update request for next iteration
		req.Messages = messages
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
```

### Combining Extended Thinking with Tool Use

```go
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
	client := api.NewClient(api.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))

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

	// Create a request with both extended thinking and tool use
	userQuery := "Calculate the area of a circle with radius 7.5 cm"
	req := models.MessageRequest{
		Model:     models.Claude37Sonnet,
		MaxTokens: 4096,
		Messages: []models.MessageParam{
			models.NewUserMessage(models.CreateTextBlock(userQuery)),
		},
		Tools:    []models.Tool{calculatorTool},
		Thinking: models.EnableThinking(3096), // Enable extended thinking
		Stream:   true,
	}

	ctx := context.Background()
	fmt.Println("\n[User]:", userQuery)

	for {
		// Create streaming request
		stream, err := client.CreateMessageStream(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n[Assistant]:")

		// Process streaming events - capture both thinking and tool use
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
						fmt.Print(event.Delta.Text)
					} else if event.Delta.Thinking != "" {
						fmt.Print(event.Delta.Thinking)
					} else if event.Delta.PartialJSON != "" {
						fmt.Print(event.Delta.PartialJSON)
					}
				}
			}
		}

		// ... Process tool results similar to the tool use example above ...
		// Check for errors and process tool results as shown in earlier examples

		// Get the final message
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
		messages := append(req.Messages, models.MessageParam{
			Role:    models.AssistantRole,
			Content: message.Content,
		})

		// Create a user message with ONLY the tool results - NOT thinking blocks
		messages = append(messages, models.MessageParam{
			Role:    models.UserRole,
			Content: toolResults,
		})

		// Update request for next iteration
		req.Messages = messages
	}
}

// handleCalculatorTool implementation as shown above...
```

### Message Roles

```go
import "github.com/joakimcarlsson/anthropic-sdk/models"

// Create a user message with text and image content
userMessage := models.NewUserMessage(
    models.CreateTextBlock("What's in this image?"),
    models.CreateImageBlock(imageSource),
)

// Create an assistant message with text content
assistantMessage := models.MessageParam{
    Role: models.AssistantRole,
    Content: []models.ContentBlock{
        models.CreateTextBlock("I can help you with that."),
    },
}

// Add messages to a request
request := models.MessageRequest{
    Model: models.Claude35SonnetV2,
    MaxTokens: 1024,
    Messages: []models.MessageParam{
        userMessage,
        assistantMessage,
        // Add more messages for multi-turn conversations
    },
}
```

### Stop Reasons

```go
import (
    "context"
    "fmt"
    
    "github.com/joakimcarlsson/anthropic-sdk"
    "github.com/joakimcarlsson/anthropic-sdk/models"
)

// Send a message request
response, err := client.CreateMessage(ctx, request)
if err != nil {
    // Handle error
}

// Check why the generation stopped
switch response.StopReason {
case models.EndTurn:
    fmt.Println("The model completed its response naturally")
case models.MaxTokens:
    fmt.Println("The response was cut off due to reaching max tokens")
case models.StopSequence:
    fmt.Println("A stop sequence was encountered")
case models.ToolUse:
    fmt.Println("The model invoked a tool")
default:
    fmt.Printf("Unknown stop reason: %s\n", response.StopReason)
}

// You can also check stop reasons in streaming responses
message := stream.Message()
fmt.Printf("Stream finished: %s\n", message.StopReason)
```

## Error Handling

This SDK provides detailed error information when API requests fail:

```go
import (
    "context"
    "fmt"
    "os"
    
    "github.com/joakimcarlsson/anthropic-sdk"
    "github.com/joakimcarlsson/anthropic-sdk/models"
)

// Attempt to create a message
response, err := client.CreateMessage(ctx, request)
if err != nil {
    // Try to cast the error to an APIError for more details
    if apiErr, ok := err.(*api.APIError); ok {
        // Access detailed error information
        fmt.Printf("API Error: %s\n", apiErr.Error())
        fmt.Printf("Status: %d\n", apiErr.StatusCode)
        fmt.Printf("Type: %s\n", apiErr.Type)
        fmt.Printf("Message: %s\n", apiErr.Message)
        
        // Check for specific error types
        if apiErr.IsRateLimitError() {
            fmt.Println("Rate limit exceeded. Implementing backoff...")
            // Implement exponential backoff
        } else if apiErr.IsBadRequestError() {
            fmt.Println("Bad request. Check your inputs.")
        } else if apiErr.IsAuthenticationError() {
            fmt.Println("Authentication failed. Check your API key.")
        } else if apiErr.IsAPIError() {
            fmt.Println("An API error occurred. Please try again later.")
        }
    } else {
        // Handle non-API errors (like network issues)
        fmt.Printf("Error: %v\n", err)
    }
    os.Exit(1)
}

// Process response
fmt.Println("Request successful!")
```

## License

MIT 