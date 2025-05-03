package models

const (
	Claude3Opus          = "claude-3-opus-20240229"
	Claude3OpusLatest    = "claude-3-opus-latest"
	Claude3Sonnet        = "claude-3-sonnet-20240229"
	Claude3Haiku         = "claude-3-haiku-20240307"
	Claude35SonnetV1     = "claude-3-5-sonnet-20240620"
	Claude35SonnetV2     = "claude-3-5-sonnet-20241022"
	Claude35SonnetLatest = "claude-3-5-sonnet-latest"
	Claude35Haiku        = "claude-3-5-haiku-20241022"
	Claude35HaikuLatest  = "claude-3-5-haiku-latest"
	Claude37Sonnet       = "claude-3-7-sonnet-20250219"
	Claude37SonnetLatest = "claude-3-7-sonnet-latest"
	Claude35Sonnet       = "claude-3-5-sonnet-20240620"
)

// ContentType defines the type of content in a message
type ContentType string

const (
	TextContentType             ContentType = "text"
	ImageContentType            ContentType = "image"
	ToolUseContentType          ContentType = "tool_use"
	ToolResultContentType       ContentType = "tool_result"
	ThinkingContentType         ContentType = "thinking"
	RedactedThinkingContentType ContentType = "redacted_thinking"
)

// Role defines the role of a message participant
type Role string

const (
	UserRole      Role = "user"
	AssistantRole Role = "assistant"
)

// StopReason defines why generation stopped
type StopReason string

const (
	EndTurn      StopReason = "end_turn"
	MaxTokens    StopReason = "max_tokens"
	StopSequence StopReason = "stop_sequence"
	ToolUse      StopReason = "tool_use"
)
