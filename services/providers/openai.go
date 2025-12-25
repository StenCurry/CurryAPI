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

// OpenAIProvider implements the ProviderClient interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// IsAvailable returns true if the provider is properly configured
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetProviderName returns the provider identifier
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// GetSupportedModels returns the list of models supported by this provider
func (p *OpenAIProvider) GetSupportedModels() []models.ModelInfo {
	isAvailable := p.IsAvailable()
	return []models.ModelInfo{
		{
			ID:            "gpt-4o",
			Name:          "GPT-4o",
			Provider:      "openai",
			ContextWindow: 128000,
			InputPrice:    2.50,
			OutputPrice:   10.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gpt-4o-mini",
			Name:          "GPT-4o Mini",
			Provider:      "openai",
			ContextWindow: 128000,
			InputPrice:    0.15,
			OutputPrice:   0.60,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gpt-4-turbo",
			Name:          "GPT-4 Turbo",
			Provider:      "openai",
			ContextWindow: 128000,
			InputPrice:    10.00,
			OutputPrice:   30.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gpt-4",
			Name:          "GPT-4",
			Provider:      "openai",
			ContextWindow: 8192,
			InputPrice:    30.00,
			OutputPrice:   60.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gpt-3.5-turbo",
			Name:          "GPT-3.5 Turbo",
			Provider:      "openai",
			ContextWindow: 16385,
			InputPrice:    0.50,
			OutputPrice:   1.50,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "o1",
			Name:          "O1",
			Provider:      "openai",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   60.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "o1-mini",
			Name:          "O1 Mini",
			Provider:      "openai",
			ContextWindow: 128000,
			InputPrice:    3.00,
			OutputPrice:   12.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "o3",
			Name:          "O3",
			Provider:      "openai",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   60.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "o3-mini",
			Name:          "O3 Mini",
			Provider:      "openai",
			ContextWindow: 128000,
			InputPrice:    3.00,
			OutputPrice:   12.00,
			IsAvailable:   isAvailable,
		},
	}
}

// ChatCompletion sends a chat request and returns a streaming channel
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI provider not available: API key not configured")
	}

	// Build the request body
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

// processStream processes the SSE stream from OpenAI
func (p *OpenAIProvider) processStream(resp *http.Response, eventChan chan<- models.StreamEvent) {
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
				// Note: OpenAI typically sends usage in a separate event or at the end
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
func (p *OpenAIProvider) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp models.ErrorResponse
	message := string(body)
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		message = errorResp.Error.Message
	}

	return p.mapErrorCode(statusCode, message)
}

// mapErrorCode maps HTTP status codes to appropriate error messages
func (p *OpenAIProvider) mapErrorCode(statusCode int, message string) error {
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
