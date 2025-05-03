package anthropic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/joakimcarlsson/anthropic-sdk/models"
	"github.com/joakimcarlsson/anthropic-sdk/streaming"
)

// Message API path
const messagesPath = "v1/messages"

// CreateMessage creates a new message
func (c *Client) CreateMessage(ctx context.Context, req models.MessageRequest) (*models.Message, error) {
	var resp models.Message
	err := c.post(ctx, messagesPath, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateMessageStream creates a new message with streaming
func (c *Client) CreateMessageStream(ctx context.Context, req models.MessageRequest) (*streaming.MessageStream, error) {
	// Ensure streaming is enabled
	req.Stream = true

	// Create custom request for streaming
	url := c.BaseURL + "/" + messagesPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", c.APIKey)
	httpReq.Header.Set("anthropic-version", c.Version)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Add body
	err = setJSONBody(httpReq, req)
	if err != nil {
		return nil, err
	}

	// Make request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making streaming request: %w", err)
	}

	// Check for error
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading error response: %w (status code: %d)", err, resp.StatusCode)
		}

		apiErr := ParseAPIError(resp.StatusCode, respData)

		// Extract request ID if present
		if requestID := resp.Header.Get("x-request-id"); requestID != "" {
			apiErr.RequestID = requestID
		}

		// Handle rate limit headers if present
		if apiErr.IsRateLimitError() {
			apiErr.RateLimitInfo = &RateLimitInfo{}
			if retryAfter := resp.Header.Get("retry-after"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					apiErr.RateLimitInfo.ResetAfter = seconds
				}
			}
			apiErr.RateLimitInfo.LimitType = resp.Header.Get("x-ratelimit-limit-type")
		}

		return nil, apiErr
	}

	// Create stream
	return streaming.NewMessageStream(resp.Body), nil
}

// CountTokens counts the tokens in a message
func (c *Client) CountTokens(ctx context.Context, req models.MessageRequest) (int, error) {
	type tokenCountResponse struct {
		InputTokens int `json:"input_tokens"`
	}

	var resp tokenCountResponse
	err := c.post(ctx, "v1/messages/count_tokens", req, &resp)
	if err != nil {
		return 0, err
	}
	return resp.InputTokens, nil
}
