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

// AnthropicProvider implements the ProviderClient interface for Anthropic
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider instance
func NewAnthropicProvider(apiKey, baseURL string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// IsAvailable returns true if the provider is properly configured
func (p *AnthropicProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetProviderName returns the provider identifier
func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

// GetSupportedModels returns the list of models supported by this provider
func (p *AnthropicProvider) GetSupportedModels() []models.ModelInfo {
	isAvailable := p.IsAvailable()
	return []models.ModelInfo{
		{
			ID:            "claude-3-5-sonnet-20241022",
			Name:          "Claude 3.5 Sonnet",
			Provider:      "anthropic",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "claude-3-5-haiku-20241022",
			Name:          "Claude 3.5 Haiku",
			Provider:      "anthropic",
			ContextWindow: 200000,
			InputPrice:    0.80,
			OutputPrice:   4.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "claude-3-opus-20240229",
			Name:          "Claude 3 Opus",
			Provider:      "anthropic",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   75.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "claude-3-sonnet-20240229",
			Name:          "Claude 3 Sonnet",
			Provider:      "anthropic",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "claude-3-haiku-20240307",
			Name:          "Claude 3 Haiku",
			Provider:      "anthropic",
			ContextWindow: 200000,
			InputPrice:    0.25,
			OutputPrice:   1.25,
			IsAvailable:   isAvailable,
		},
	}
}

// AnthropicMessage represents a message in Anthropic's format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents the request body for Anthropic API
type AnthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []AnthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Stream      bool                `json:"stream"`
	System      string              `json:"system,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
}

// AnthropicStreamEvent represents different event types from Anthropic's streaming API
type AnthropicStreamEvent struct {
	Type         string                    `json:"type"`
	Message      *AnthropicMessageResponse `json:"message,omitempty"`
	Index        int                       `json:"index,omitempty"`
	ContentBlock *AnthropicContentBlock    `json:"content_block,omitempty"`
	Delta        *AnthropicDelta           `json:"delta,omitempty"`
	Usage        *AnthropicUsage           `json:"usage,omitempty"`
}

// AnthropicMessageResponse represents the message in Anthropic's response
type AnthropicMessageResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []AnthropicContent  `json:"content"`
	Model        string              `json:"model"`
	StopReason   *string             `json:"stop_reason"`
	StopSequence *string             `json:"stop_sequence"`
	Usage        *AnthropicUsage     `json:"usage"`
}

// AnthropicContent represents content in Anthropic's response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicContentBlock represents a content block in streaming
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicDelta represents delta updates in streaming
type AnthropicDelta struct {
	Type         string          `json:"type"`
	Text         string          `json:"text,omitempty"`
	StopReason   *string         `json:"stop_reason,omitempty"`
	StopSequence *string         `json:"stop_sequence,omitempty"`
}

// AnthropicUsage represents token usage information
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// convertToAnthropicFormat converts OpenAI-style messages to Anthropic format
func (p *AnthropicProvider) convertToAnthropicFormat(messages []models.Message) ([]AnthropicMessage, string, error) {
	var anthropicMessages []AnthropicMessage
	var systemPrompt string

	for _, msg := range messages {
		// Extract system prompt separately
		if msg.Role == "system" {
			content := ""
			switch v := msg.Content.(type) {
			case string:
				content = v
			case []interface{}:
				// Handle array content
				for _, part := range v {
					if partMap, ok := part.(map[string]interface{}); ok {
						if text, ok := partMap["text"].(string); ok {
							content += text
						}
					}
				}
			}
			if systemPrompt != "" {
				systemPrompt += "\n"
			}
			systemPrompt += content
			continue
		}

		// Convert user/assistant messages
		if msg.Role == "user" || msg.Role == "assistant" {
			content := ""
			switch v := msg.Content.(type) {
			case string:
				content = v
			case []interface{}:
				// Handle array content
				for _, part := range v {
					if partMap, ok := part.(map[string]interface{}); ok {
						if text, ok := partMap["text"].(string); ok {
							content += text
						}
					}
				}
			}

			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    msg.Role,
				Content: content,
			})
		}
	}

	return anthropicMessages, systemPrompt, nil
}

// ChatCompletion sends a chat request and returns a streaming channel
func (p *AnthropicProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("Anthropic provider not available: API key not configured")
	}

	// Convert messages to Anthropic format
	anthropicMessages, systemPrompt, err := p.convertToAnthropicFormat(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Build the request body
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096 // Anthropic requires max_tokens
	}

	requestBody := AnthropicRequest{
		Model:     req.Model,
		Messages:  anthropicMessages,
		MaxTokens: maxTokens,
		Stream:    true,
		System:    systemPrompt,
	}

	if req.Temperature > 0 {
		requestBody.Temperature = req.Temperature
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := p.baseURL + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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

// processStream processes the SSE stream from Anthropic
func (p *AnthropicProvider) processStream(resp *http.Response, eventChan chan<- models.StreamEvent) {
	defer close(eventChan)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var totalUsage *models.TokenUsage

	// Send start event
	eventChan <- models.StreamEvent{
		Type: "start",
	}

	var currentEvent string
	var currentData string

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines (they separate events)
		if line == "" {
			if currentEvent != "" && currentData != "" {
				p.processAnthropicEvent(currentEvent, currentData, eventChan, &totalUsage)
				currentEvent = ""
				currentData = ""
			}
			continue
		}

		// Parse event type
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		// Parse data
		if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
			continue
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

// processAnthropicEvent processes a single Anthropic SSE event
func (p *AnthropicProvider) processAnthropicEvent(eventType, data string, eventChan chan<- models.StreamEvent, totalUsage **models.TokenUsage) {
	var streamEvent AnthropicStreamEvent
	if err := json.Unmarshal([]byte(data), &streamEvent); err != nil {
		eventChan <- models.StreamEvent{
			Type:  "error",
			Error: fmt.Sprintf("failed to parse stream event: %v", err),
		}
		return
	}

	switch eventType {
	case "message_start":
		// Extract initial token usage (input tokens)
		if streamEvent.Message != nil && streamEvent.Message.Usage != nil {
			if *totalUsage == nil {
				*totalUsage = &models.TokenUsage{}
			}
			(*totalUsage).PromptTokens = streamEvent.Message.Usage.InputTokens
		}

	case "content_block_start":
		// Content block started, no action needed

	case "content_block_delta":
		// Send content delta
		if streamEvent.Delta != nil && streamEvent.Delta.Text != "" {
			eventChan <- models.StreamEvent{
				Type:    "content",
				Content: streamEvent.Delta.Text,
			}
		}

	case "content_block_stop":
		// Content block stopped, no action needed

	case "message_delta":
		// Extract output token usage
		if streamEvent.Usage != nil {
			if *totalUsage == nil {
				*totalUsage = &models.TokenUsage{}
			}
			(*totalUsage).CompletionTokens = streamEvent.Usage.OutputTokens
			(*totalUsage).TotalTokens = (*totalUsage).PromptTokens + (*totalUsage).CompletionTokens
		}

	case "message_stop":
		// Message completed, handled in main loop

	case "error":
		eventChan <- models.StreamEvent{
			Type:  "error",
			Error: "Anthropic API error",
		}
	}
}

// handleErrorResponse converts HTTP error responses to appropriate errors
func (p *AnthropicProvider) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	message := string(body)
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		message = errorResp.Error.Message
	}

	return p.mapErrorCode(statusCode, message)
}

// mapErrorCode maps HTTP status codes to appropriate error messages
func (p *AnthropicProvider) mapErrorCode(statusCode int, message string) error {
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
