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

// GoogleProvider implements the ProviderClient interface for Google AI
type GoogleProvider struct {
	apiKey string
	client *http.Client
}

// NewGoogleProvider creates a new Google AI provider instance
func NewGoogleProvider(apiKey string) *GoogleProvider {
	return &GoogleProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// IsAvailable returns true if the provider is properly configured
func (p *GoogleProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetProviderName returns the provider identifier
func (p *GoogleProvider) GetProviderName() string {
	return "google"
}

// GetSupportedModels returns the list of models supported by this provider
func (p *GoogleProvider) GetSupportedModels() []models.ModelInfo {
	isAvailable := p.IsAvailable()
	return []models.ModelInfo{
		{
			ID:            "gemini-1.5-pro",
			Name:          "Gemini 1.5 Pro",
			Provider:      "google",
			ContextWindow: 2097152, // 2M tokens
			InputPrice:    1.25,
			OutputPrice:   5.00,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gemini-1.5-flash",
			Name:          "Gemini 1.5 Flash",
			Provider:      "google",
			ContextWindow: 1048576, // 1M tokens
			InputPrice:    0.075,
			OutputPrice:   0.30,
			IsAvailable:   isAvailable,
		},
		{
			ID:            "gemini-pro",
			Name:          "Gemini Pro",
			Provider:      "google",
			ContextWindow: 32768,
			InputPrice:    0.50,
			OutputPrice:   1.50,
			IsAvailable:   isAvailable,
		},
	}
}

// GoogleContent represents a content part in Google's format
type GoogleContent struct {
	Role  string        `json:"role"`
	Parts []GooglePart  `json:"parts"`
}

// GooglePart represents a part of the content
type GooglePart struct {
	Text string `json:"text"`
}

// GoogleRequest represents the request body for Google AI API
type GoogleRequest struct {
	Contents         []GoogleContent           `json:"contents"`
	GenerationConfig *GoogleGenerationConfig   `json:"generationConfig,omitempty"`
}

// GoogleGenerationConfig represents generation configuration
type GoogleGenerationConfig struct {
	Temperature  float64 `json:"temperature,omitempty"`
	MaxOutputTokens int  `json:"maxOutputTokens,omitempty"`
}

// GoogleStreamResponse represents a streaming response from Google AI
type GoogleStreamResponse struct {
	Candidates    []GoogleCandidate    `json:"candidates,omitempty"`
	UsageMetadata *GoogleUsageMetadata `json:"usageMetadata,omitempty"`
}

// GoogleCandidate represents a candidate response
type GoogleCandidate struct {
	Content      GoogleContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

// GoogleUsageMetadata represents token usage information
type GoogleUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// convertToGoogleFormat converts OpenAI-style messages to Google format
func (p *GoogleProvider) convertToGoogleFormat(messages []models.Message) ([]GoogleContent, error) {
	var googleContents []GoogleContent

	for _, msg := range messages {
		// Google uses "user" and "model" roles (not "assistant")
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		// System messages are typically prepended to the first user message in Google
		if role == "system" {
			role = "user"
		}

		// Extract text content
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

		googleContents = append(googleContents, GoogleContent{
			Role: role,
			Parts: []GooglePart{
				{Text: content},
			},
		})
	}

	return googleContents, nil
}

// ChatCompletion sends a chat request and returns a streaming channel
func (p *GoogleProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("Google provider not available: API key not configured")
	}

	// Convert messages to Google format
	googleContents, err := p.convertToGoogleFormat(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Build the request body
	requestBody := GoogleRequest{
		Contents: googleContents,
	}

	// Add generation config if needed
	if req.Temperature > 0 || req.MaxTokens > 0 {
		requestBody.GenerationConfig = &GoogleGenerationConfig{}
		if req.Temperature > 0 {
			requestBody.GenerationConfig.Temperature = req.Temperature
		}
		if req.MaxTokens > 0 {
			requestBody.GenerationConfig.MaxOutputTokens = req.MaxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with API key as query parameter
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s&alt=sse",
		req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

// processStream processes the SSE stream from Google AI
func (p *GoogleProvider) processStream(resp *http.Response, eventChan chan<- models.StreamEvent) {
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

		// Parse JSON
		var streamResp GoogleStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			eventChan <- models.StreamEvent{
				Type:  "error",
				Error: fmt.Sprintf("failed to parse stream response: %v", err),
			}
			return
		}

		// Process candidates
		if len(streamResp.Candidates) > 0 {
			candidate := streamResp.Candidates[0]

			// Send content
			if len(candidate.Content.Parts) > 0 {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						eventChan <- models.StreamEvent{
							Type:    "content",
							Content: part.Text,
						}
					}
				}
			}
		}

		// Extract usage metadata
		if streamResp.UsageMetadata != nil {
			totalUsage = &models.TokenUsage{
				PromptTokens:     streamResp.UsageMetadata.PromptTokenCount,
				CompletionTokens: streamResp.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      streamResp.UsageMetadata.TotalTokenCount,
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
func (p *GoogleProvider) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	message := string(body)
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		message = errorResp.Error.Message
	}

	return p.mapErrorCode(statusCode, message)
}

// mapErrorCode maps HTTP status codes to appropriate error messages
func (p *GoogleProvider) mapErrorCode(statusCode int, message string) error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
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
