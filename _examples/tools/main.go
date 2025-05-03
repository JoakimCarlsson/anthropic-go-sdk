package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joakimcarlsson/anthropic-sdk"
	"github.com/joakimcarlsson/anthropic-sdk/models"
)

func main() {
	// Create client
	client := anthropic.NewClient()

	// Define tools
	weatherTool := models.NewTool(
		"get_weather",
		"Get current weather for a location",
		models.SimpleJSONSchema(
			map[string]models.Property{
				"location": models.NewProperty("string", "The city and state/country"),
				"unit": models.NewEnumProperty(
					"Temperature unit",
					[]string{"celsius", "fahrenheit"},
				),
			},
			[]string{"location"},
		),
	)

	timeTool := models.NewTool(
		"get_current_time",
		"Get the current time for a location",
		models.SimpleJSONSchema(
			map[string]models.Property{
				"location": models.NewProperty("string", "The city and state/country"),
			},
			[]string{"location"},
		),
	)

	// Create conversation history
	messages := []models.MessageParam{
		models.NewUserMessage(models.CreateTextBlock("What's the weather like in New York? And what time is it there now?")),
	}

	// Create message request
	req := models.MessageRequest{
		Model:     models.Claude35SonnetV2,
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     []models.Tool{weatherTool, timeTool},
	}

	// Loop to handle tool use
	ctx := context.Background()
	for {
		// Call API
		resp, err := client.CreateMessage(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Print assistant's response
		fmt.Println("\n[Assistant]:")
		for _, block := range resp.Content {
			if block.TextContent != nil {
				fmt.Println(block.TextContent.Text)
			} else if block.ToolUseContent != nil {
				fmt.Printf("[Using tool: %s]\n", block.ToolUseContent.Name)
			}
		}

		// Check if any tools were used
		toolResults := []models.ContentBlock{}
		for _, block := range resp.Content {
			if block.ToolUseContent != nil {
				// Handle tool use
				toolUseID := block.ToolUseContent.ID
				toolName := block.ToolUseContent.Name
				toolInput := block.ToolUseContent.Input

				// Process based on tool type
				var result string
				switch toolName {
				case "get_weather":
					result = handleWeatherTool(toolInput)
				case "get_current_time":
					result = handleTimeTool(toolInput)
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

		// Add assistant's response and tool results to messages
		messages = append(messages, models.MessageParam{
			Role:    models.AssistantRole,
			Content: resp.Content,
		})
		messages = append(messages, models.MessageParam{
			Role:    models.UserRole,
			Content: toolResults,
		})

		// Update request
		req.Messages = messages
	}
}

func handleWeatherTool(inputData interface{}) string {
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
		return "Error: Received empty JSON input for weather tool"
	}

	var input struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	// Check if location is empty
	if input.Location == "" {
		return "Error: No location provided"
	}

	// Default to celsius if not specified
	unit := input.Unit
	if unit == "" {
		unit = "celsius"
	}

	// Mock weather data
	temperature := 22
	condition := "Partly Cloudy"
	if unit == "fahrenheit" {
		temperature = temperature*9/5 + 32
	}

	// Return formatted result
	return fmt.Sprintf(
		"Weather in %s: %d°%s, %s",
		input.Location,
		temperature,
		unitSymbol(unit),
		condition,
	)
}

func handleTimeTool(inputData interface{}) string {
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
		return "Error: Received empty JSON input for time tool"
	}

	var input struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	// Check if location is empty
	if input.Location == "" {
		return "Error: No location provided"
	}

	// Mock time data (just using local time)
	now := time.Now().Format("3:04 PM MST")

	// Return formatted result
	return fmt.Sprintf(
		"Current time in %s: %s",
		input.Location,
		now,
	)
}

func unitSymbol(unit string) string {
	if unit == "celsius" {
		return "C"
	}
	return "F"
}
