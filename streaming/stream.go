package streaming

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/joakimcarlsson/anthropic-sdk/models"
)

// EventType represents the type of streaming event
type EventType string

const (
	MessageStartEvent      EventType = "message_start"
	ContentBlockStartEvent EventType = "content_block_start"
	ContentBlockDeltaEvent EventType = "content_block_delta"
	ContentBlockStopEvent  EventType = "content_block_stop"
	MessageDeltaEvent      EventType = "message_delta"
	MessageStopEvent       EventType = "message_stop"
)

// Event represents a streaming event
type Event struct {
	Type         EventType            `json:"type"`
	Message      *models.Message      `json:"message,omitempty"`
	StopReason   *models.StopReason   `json:"stop_reason,omitempty"`
	Index        *int                 `json:"index,omitempty"`
	ContentBlock *models.ContentBlock `json:"content_block,omitempty"`
	Delta        *Delta               `json:"delta,omitempty"`
	Usage        *models.Usage        `json:"usage,omitempty"`
}

// Delta represents a delta update in a streaming event
type Delta struct {
	Type        string `json:"type,omitempty"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	Thinking    string `json:"thinking,omitempty"`
	Signature   string `json:"signature,omitempty"`
}

// MessageStream handles streaming responses from the Claude API
type MessageStream struct {
	reader       *bufio.Reader
	currentEvent *Event
	err          error
	message      *models.Message
	jsonBuffers  map[int]string
}

// NewMessageStream creates a new message stream from a reader
func NewMessageStream(reader io.Reader) *MessageStream {
	return &MessageStream{
		reader:      bufio.NewReader(reader),
		message:     &models.Message{},
		jsonBuffers: make(map[int]string),
	}
}

// Next advances the stream to the next event
func (s *MessageStream) Next() bool {
	if s.err != nil {
		return false
	}

	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return false
		}
		s.err = fmt.Errorf("error reading stream: %w", err)
		return false
	}

	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return s.Next()
	}

	prefix := []byte("data: ")
	if !bytes.HasPrefix(line, prefix) {
		return s.Next()
	}

	data := line[len(prefix):]
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		s.err = fmt.Errorf("error parsing event: %w", err)
		return false
	}

	s.currentEvent = &event
	s.updateMessage(&event)

	return true
}

// Current returns the current event
func (s *MessageStream) Current() *Event {
	return s.currentEvent
}

// Err returns any error that occurred during streaming
func (s *MessageStream) Err() error {
	return s.err
}

// Message returns the accumulated message
func (s *MessageStream) Message() *models.Message {
	return s.message
}

// updateMessage updates the accumulated message with the current event
func (s *MessageStream) updateMessage(event *Event) {
	switch event.Type {
	case MessageStartEvent:
		if event.Message != nil {
			s.message.ID = event.Message.ID
			s.message.Role = event.Message.Role
			s.message.Model = event.Message.Model
		}
	case ContentBlockStartEvent:
		if event.ContentBlock != nil && event.Index != nil {
			idx := *event.Index
			for len(s.message.Content) <= idx {
				s.message.Content = append(s.message.Content, models.ContentBlock{})
			}
			s.message.Content[idx] = *event.ContentBlock

			if event.ContentBlock.TextContent != nil && event.ContentBlock.TextContent.Text == "" {
			}

			if event.ContentBlock.ToolUseContent != nil {
				s.jsonBuffers[idx] = ""
			}
		}
	case ContentBlockDeltaEvent:
		if event.Delta != nil && event.Index != nil {
			idx := *event.Index
			if idx < len(s.message.Content) {
				if event.Delta.Type == "text_delta" {
					if s.message.Content[idx].TextContent != nil {
						s.message.Content[idx].TextContent.Text += event.Delta.Text
					}
				} else if event.Delta.Type == "input_json_delta" {
					if s.message.Content[idx].ToolUseContent != nil {
						s.jsonBuffers[idx] += event.Delta.PartialJSON

						jsonStr := s.jsonBuffers[idx]
						if strings.HasPrefix(jsonStr, "{") && strings.HasSuffix(jsonStr, "}") {
							var inputObj map[string]interface{}
							if err := json.Unmarshal([]byte(jsonStr), &inputObj); err == nil {
								s.message.Content[idx].ToolUseContent.Input = inputObj
							}
						}
					}
				} else if event.Delta.Type == "thinking_delta" {
					if s.message.Content[idx].ThinkingContent != nil {
						s.message.Content[idx].ThinkingContent.Thinking += event.Delta.Thinking
					}
				} else if event.Delta.Type == "signature_delta" {
					if s.message.Content[idx].ThinkingContent != nil {
						s.message.Content[idx].ThinkingContent.Signature = event.Delta.Signature
					}
				}
			}
		}
	case ContentBlockStopEvent:
		if event.Index != nil {
			idx := *event.Index

			if idx < len(s.message.Content) && s.message.Content[idx].ToolUseContent != nil {
				jsonStr := s.jsonBuffers[idx]
				if strings.HasPrefix(jsonStr, "{") && strings.HasSuffix(jsonStr, "}") {
					var inputObj map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &inputObj); err == nil {
						s.message.Content[idx].ToolUseContent.Input = inputObj
					}
				}
			}
		}
	case MessageStopEvent:
		if event.StopReason != nil {
			s.message.StopReason = *event.StopReason
		}
		if event.Usage != nil {
			s.message.Usage = *event.Usage
		}
	}
}
