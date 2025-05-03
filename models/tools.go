package models

// Tool represents a tool that can be used by Claude
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema represents the schema for a tool's input
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property represents a property in an input schema
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolChoice represents how tools should be used by Claude
type ToolChoice struct {
	Type                   string `json:"type"`
	Name                   string `json:"name,omitempty"`
	DisableParallelToolUse bool   `json:"disable_parallel_tool_use,omitempty"`
}

// NewTool creates a new tool
func NewTool(name string, description string, schema InputSchema) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}
}

// SimpleJSONSchema creates a simple JSON schema for object properties
func SimpleJSONSchema(properties map[string]Property, required []string) InputSchema {
	return InputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// NewProperty creates a new property
func NewProperty(propertyType, description string) Property {
	return Property{
		Type:        propertyType,
		Description: description,
	}
}

// NewEnumProperty creates a new enum property
func NewEnumProperty(description string, values []string) Property {
	return Property{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
}

// AutoToolChoice creates an automatic tool choice
func AutoToolChoice() ToolChoice {
	return ToolChoice{
		Type: "auto",
	}
}

// SpecificToolChoice creates a tool choice for a specific tool
func SpecificToolChoice(name string, disableParallel bool) ToolChoice {
	return ToolChoice{
		Type:                   "tool",
		Name:                   name,
		DisableParallelToolUse: disableParallel,
	}
}

// NoToolChoice creates a tool choice that disables tool use
func NoToolChoice() ToolChoice {
	return ToolChoice{
		Type: "none",
	}
}

// CreateToolUseBlock creates a new tool use content block
func CreateToolUseBlock(id string, name string, input interface{}) ContentBlock {
	return ContentBlock{
		ToolUseContent: &ToolUseBlock{
			Type:  ToolUseContentType,
			ID:    id,
			Name:  name,
			Input: input,
		},
	}
}
