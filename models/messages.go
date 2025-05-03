package models

import (
	"encoding/json"
)

// Message represents a message in a conversation
type Message struct {
	ID           string         `json:"id"`
	Role         Role           `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   StopReason     `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// MessageParam represents an input message
type MessageParam struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// TextBlock represents a text content block
type TextBlock struct {
	Type ContentType `json:"type"`
	Text string      `json:"text"`
}

// ImageBlock represents an image content block
type ImageBlock struct {
	Type   ContentType `json:"type"`
	Source ImageSource `json:"source"`
}

// ToolUseBlock represents a tool use content block
type ToolUseBlock struct {
	Type  ContentType `json:"type"`
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

// ToolResultBlock represents a tool result content block
type ToolResultBlock struct {
	Type      ContentType `json:"type"`
	ToolUseID string      `json:"tool_use_id"`
	Content   string      `json:"content"`
	IsError   bool        `json:"is_error,omitempty"`
}

// ThinkingBlock represents a thinking content block
type ThinkingBlock struct {
	Type      ContentType `json:"type"`
	Thinking  string      `json:"thinking"`
	Signature string      `json:"signature"`
}

// RedactedThinkingBlock represents a redacted thinking content block
type RedactedThinkingBlock struct {
	Type ContentType `json:"type"`
	Data string      `json:"data"`
}

// ContentBlock represents a block of content in a message
type ContentBlock struct {
	TextContent             *TextBlock             `json:"-"`
	ImageContent            *ImageBlock            `json:"-"`
	ToolUseContent          *ToolUseBlock          `json:"-"`
	ToolResultContent       *ToolResultBlock       `json:"-"`
	ThinkingContent         *ThinkingBlock         `json:"-"`
	RedactedThinkingContent *RedactedThinkingBlock `json:"-"`
}

// MarshalJSON implements the json.Marshaler interface
func (c ContentBlock) MarshalJSON() ([]byte, error) {
	if c.TextContent != nil {
		return json.Marshal(c.TextContent)
	}
	if c.ImageContent != nil {
		return json.Marshal(c.ImageContent)
	}
	if c.ToolUseContent != nil {
		return json.Marshal(c.ToolUseContent)
	}
	if c.ToolResultContent != nil {
		return json.Marshal(c.ToolResultContent)
	}
	if c.ThinkingContent != nil {
		return json.Marshal(c.ThinkingContent)
	}
	if c.RedactedThinkingContent != nil {
		return json.Marshal(c.RedactedThinkingContent)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *ContentBlock) UnmarshalJSON(data []byte) error {
	var typeCheck struct {
		Type ContentType `json:"type"`
	}
	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return err
	}

	switch typeCheck.Type {
	case TextContentType:
		var textBlock TextBlock
		if err := json.Unmarshal(data, &textBlock); err != nil {
			return err
		}
		c.TextContent = &textBlock
	case ImageContentType:
		var imageBlock ImageBlock
		if err := json.Unmarshal(data, &imageBlock); err != nil {
			return err
		}
		c.ImageContent = &imageBlock
	case ToolUseContentType:
		var toolUseBlock ToolUseBlock
		if err := json.Unmarshal(data, &toolUseBlock); err != nil {
			return err
		}
		c.ToolUseContent = &toolUseBlock
	case ToolResultContentType:
		var toolResultBlock ToolResultBlock
		if err := json.Unmarshal(data, &toolResultBlock); err != nil {
			return err
		}
		c.ToolResultContent = &toolResultBlock
	case ThinkingContentType:
		var thinkingBlock ThinkingBlock
		if err := json.Unmarshal(data, &thinkingBlock); err != nil {
			return err
		}
		c.ThinkingContent = &thinkingBlock
	case RedactedThinkingContentType:
		var redactedThinkingBlock RedactedThinkingBlock
		if err := json.Unmarshal(data, &redactedThinkingBlock); err != nil {
			return err
		}
		c.RedactedThinkingContent = &redactedThinkingBlock
	}

	return nil
}

// CreateTextBlock creates a new text content block
func CreateTextBlock(text string) ContentBlock {
	return ContentBlock{
		TextContent: &TextBlock{
			Type: TextContentType,
			Text: text,
		},
	}
}

// CreateToolResultBlock creates a new tool result content block
func CreateToolResultBlock(toolUseID string, content string, isError bool) ContentBlock {
	return ContentBlock{
		ToolResultContent: &ToolResultBlock{
			Type:      ToolResultContentType,
			ToolUseID: toolUseID,
			Content:   content,
			IsError:   isError,
		},
	}
}

// MessageRequest represents a request to create a message
type MessageRequest struct {
	Model         string          `json:"model"`
	Messages      []MessageParam  `json:"messages"`
	System        string          `json:"system,omitempty"`
	MaxTokens     int             `json:"max_tokens"`
	Temperature   *float64        `json:"temperature,omitempty"`
	TopP          *float64        `json:"top_p,omitempty"`
	TopK          *int            `json:"top_k,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Stream        bool            `json:"stream,omitempty"`
	Tools         []Tool          `json:"tools,omitempty"`
	ToolChoice    *ToolChoice     `json:"tool_choice,omitempty"`
	Thinking      *ThinkingConfig `json:"thinking,omitempty"`
}

// ThinkingConfig represents the configuration for extended thinking
type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// EnableThinking creates a new thinking configuration for extended thinking
func EnableThinking(budgetTokens int) *ThinkingConfig {
	return &ThinkingConfig{
		Type:         "enabled",
		BudgetTokens: budgetTokens,
	}
}

// Usage represents token usage statistics for an API call
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewUserMessage creates a new user message
func NewUserMessage(content ...ContentBlock) MessageParam {
	return MessageParam{
		Role:    UserRole,
		Content: content,
	}
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(content ...ContentBlock) MessageParam {
	return MessageParam{
		Role:    AssistantRole,
		Content: content,
	}
}
