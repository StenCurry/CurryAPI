package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Curry2API-go/models"
)

// DeepSeekProvider implements the ProviderClient interface for DeepSeek
type DeepSeekProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewDeepSeekProvider creates a new DeepSeek provider instance
func NewDeepSeekProvider(apiKey, baseURL string) *DeepSeekProvider {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	return &DeepSeekProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// IsAvailable returns true if the provider is properly configured
func (p *DeepSeekProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetProviderName returns the provider identifier
func (p *DeepSeekProvider) GetProviderName() string {
	return "deepseek"
}

// GetSupportedModels returns the list of models supported by this provider
func (p *DeepSeekProvider) GetSupportedModels() []models.ModelInfo {
	isAvailable := p.IsAvailable()
	return []models.ModelInfo{
		{
			ID:            "deepseek-chat",
			Name:          "DeepSeek Chat",
			Provider:      "deepseek",
			ContextWindow: 64000,
			InputPrice:    0.14,
			OutputPrice:   0.28,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "deepseek-coder",
			Name:          "DeepSeek Coder",
			Provider:      "deepseek",
			ContextWindow: 64000,
			InputPrice:    0.14,
			OutputPrice:   0.28,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "deepseek-reasoner",
			Name:          "DeepSeek Reasoner",
			Provider:      "deepseek",
			ContextWindow: 64000,
			InputPrice:    0.55,
			OutputPrice:   2.19,
			IsAvailable:   isAvailable,
		},
	}
}

// ChatCompletion sends a chat request and returns a streaming channel
func (p *DeepSeekProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("DeepSeek provider not available: API key not configured")
	}

	// Build the request body (OpenAI-compatible format)
	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}

	if req.MaxTokens > 0 {
		requestBody["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, p.handleErrorResponse(resp.StatusCode, body)
	}

	// Create channel for streaming events
	eventChan := make(chan models.StreamEvent)

	// Start goroutine to process streaming response
	go p.processStream(resp, eventChan)

	return eventChan, nil
}

// processStream processes the SSE stream from DeepSeek
func (p *DeepSeekProvider) processStream(resp *http.Response, eventChan chan<- models.StreamEvent) {
	defer close(eventChan)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var totalUsage *models.TokenUsage

	// Send start event
	eventChan <- models.StreamEvent{
		Type: "start",
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for "data: " prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract data after "data: " prefix
		data := strings.TrimPrefix(line, "data: ")

		// Check for [DONE] marker
		if data == "[DONE]" {
			break
		}

		// Parse JSON
		var streamResp models.ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			eventChan <- models.StreamEvent{
				Type:  "error",
				Error: fmt.Sprintf("failed to parse stream response: %v", err),
			}
			return
		}

		// Process choices
		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			// Send content delta
			if choice.Delta.Content != "" {
				eventChan <- models.StreamEvent{
					Type:    "content",
					Content: choice.Delta.Content,
				}
			}

			// Check for finish reason (indicates completion)
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				// DeepSeek typically sends usage in a separate event or at the end
				// We'll try to extract it if available
			}
		}
	}

	if err := scanner.Err(); err != nil {
		eventChan <- models.StreamEvent{
			Type:  "error",
			Error: fmt.Sprintf("stream reading error: %v", err),
		}
		return
	}

	// Send usage event if we have token information
	if totalUsage != nil {
		eventChan <- models.StreamEvent{
			Type:   "usage",
			Tokens: totalUsage,
		}
	}

	// Send done event
	eventChan <- models.StreamEvent{
		Type: "done",
	}
}

// handleErrorResponse converts HTTP error responses to appropriate errors
func (p *DeepSeekProvider) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp models.ErrorResponse
	message := string(body)
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		message = errorResp.Error.Message
	}

	return p.mapErrorCode(statusCode, message)
}

// mapErrorCode maps HTTP status codes to appropriate error messages
func (p *DeepSeekProvider) mapErrorCode(statusCode int, message string) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("INVALID_API_KEY: API key is invalid or expired")
	case http.StatusTooManyRequests:
		return fmt.Errorf("RATE_LIMITED: Rate limit exceeded, please try again later")
	case http.StatusBadRequest:
		// Check if it's a context length error
		lowerMsg := strings.ToLower(message)
		if strings.Contains(lowerMsg, "context") || 
		   strings.Contains(lowerMsg, "token") ||
		   strings.Contains(lowerMsg, "maximum") ||
		   strings.Contains(lowerMsg, "length") {
			return fmt.Errorf("CONTEXT_TOO_LONG: %s", message)
		}
		return fmt.Errorf("BAD_REQUEST: %s", message)
	default:
		if statusCode >= 500 {
			return fmt.Errorf("PROVIDER_ERROR: AI service temporarily unavailable")
		}
		return fmt.Errorf("UNKNOWN_ERROR: %s", message)
	}
}
