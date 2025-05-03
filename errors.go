package anthropic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError represents an error response from the Anthropic API
type APIError struct {
	Type          string            `json:"type"`
	Message       string            `json:"message"`
	Code          string            `json:"code,omitempty"`
	Param         string            `json:"param,omitempty"`
	StatusCode    int               `json:"-"`
	RawResponse   string            `json:"-"`
	RequestID     string            `json:"request_id,omitempty"`
	RateLimitInfo *RateLimitInfo    `json:"-"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	ResetAfter int    `json:"-"`
	LimitType  string `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	var parts []string

	mainError := "api error"
	if e.Type != "" {
		mainError = fmt.Sprintf("%s: %s", mainError, e.Type)
	}
	if e.Code != "" {
		mainError = fmt.Sprintf("%s (%s)", mainError, e.Code)
	}
	if e.Message != "" {
		mainError = fmt.Sprintf("%s: %s", mainError, e.Message)
	}
	parts = append(parts, mainError)

	if e.StatusCode > 0 {
		parts = append(parts, fmt.Sprintf("HTTP Status: %d", e.StatusCode))
	}

	if e.RequestID != "" {
		parts = append(parts, fmt.Sprintf("Request ID: %s", e.RequestID))
	}

	if e.Param != "" {
		parts = append(parts, fmt.Sprintf("Invalid parameter: %s", e.Param))
	}

	if e.RateLimitInfo != nil && e.RateLimitInfo.ResetAfter > 0 {
		parts = append(parts, fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", e.RateLimitInfo.ResetAfter))
	}

	if len(e.Metadata) > 0 {
		metadataStr := "Additional info:"
		for k, v := range e.Metadata {
			metadataStr += fmt.Sprintf(" %s=%s", k, v)
		}
		parts = append(parts, metadataStr)
	}

	if e.RawResponse != "" {
		maxLen := 500
		respStr := e.RawResponse
		if len(respStr) > maxLen {
			respStr = respStr[:maxLen] + "..."
		}
		parts = append(parts, fmt.Sprintf("Raw response: %s", respStr))
	}

	return strings.Join(parts, ". ")
}

// ParseAPIError attempts to parse an API error from a JSON response
func ParseAPIError(statusCode int, data []byte) *APIError {
	var apiErr APIError
	apiErr.StatusCode = statusCode
	apiErr.RawResponse = string(data)

	var anthropicResp struct {
		Type  string    `json:"type"`
		Error *APIError `json:"error"`
	}

	if err := json.Unmarshal(data, &anthropicResp); err == nil && anthropicResp.Error != nil {
		anthropicResp.Error.StatusCode = statusCode
		anthropicResp.Error.RawResponse = string(data)
		return anthropicResp.Error
	}

	if err := json.Unmarshal(data, &apiErr); err != nil {
		return &APIError{
			Type:        "parse_error",
			Message:     fmt.Sprintf("Failed to parse error response: %v", err),
			StatusCode:  statusCode,
			RawResponse: string(data),
		}
	}

	if apiErr.IsRateLimitError() {
		apiErr.RateLimitInfo = &RateLimitInfo{}
	}

	return &apiErr
}

// IsRateLimitError returns true if the error is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	return e.Type == "rate_limit_error"
}

// IsInvalidRequestError returns true if the error is an invalid request error
func (e *APIError) IsInvalidRequestError() bool {
	return e.Type == "invalid_request_error"
}

// IsAuthenticationError returns true if the error is an authentication error
func (e *APIError) IsAuthenticationError() bool {
	return e.Type == "authentication_error"
}

// IsInternalError returns true if the error is an internal error
func (e *APIError) IsInternalError() bool {
	return e.Type == "internal_error"
}

// IsPermissionError returns true if the error is a permission error
func (e *APIError) IsPermissionError() bool {
	return e.Type == "permission_error"
}

// IsModelNotAvailableError returns true if the error indicates the requested model is not available
func (e *APIError) IsModelNotAvailableError() bool {
	return e.Code == "model_not_available" || strings.Contains(e.Message, "model not available")
}
