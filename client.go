package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	DefaultBaseURL = "https://api.anthropic.com"
	DefaultVersion = "2023-06-01"
	DefaultTimeout = 120 * time.Second
)

// Client provides a client to the Anthropic API
type Client struct {
	BaseURL    string
	APIKey     string
	Version    string
	HTTPClient *http.Client
}

// ClientOption is a function that modifies a Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithAPIKey sets the API key for the client
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.APIKey = apiKey
	}
}

// WithVersion sets the API version for the client
func WithVersion(version string) ClientOption {
	return func(c *Client) {
		c.Version = version
	}
}

// WithHTTPClient sets the HTTP client for the API client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewClient creates a new Anthropic API client
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		BaseURL:    DefaultBaseURL,
		Version:    DefaultVersion,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}

	for _, option := range options {
		option(client)
	}

	if client.APIKey == "" {
		client.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	return client
}

// request makes an HTTP request to the Anthropic API
func (c *Client) request(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	var body io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("anthropic-version", c.Version)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := ParseAPIError(resp.StatusCode, respData)

		if requestID := resp.Header.Get("x-request-id"); requestID != "" {
			apiErr.RequestID = requestID
		}

		if apiErr.IsRateLimitError() {
			apiErr.RateLimitInfo = &RateLimitInfo{}
			if retryAfter := resp.Header.Get("retry-after"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					apiErr.RateLimitInfo.ResetAfter = seconds
				}
			}
			apiErr.RateLimitInfo.LimitType = resp.Header.Get("x-ratelimit-limit-type")
		}

		return apiErr
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("error unmarshaling response: %w", err)
		}
	}

	return nil
}

// post makes a POST request to the Anthropic API
func (c *Client) post(ctx context.Context, path string, reqBody, respBody interface{}) error {
	return c.request(ctx, http.MethodPost, path, reqBody, respBody)
}
