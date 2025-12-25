package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"Curry2API-go/middleware"
	"Curry2API-go/models"
)

// CursorServiceInterface defines the interface for the Cursor service
// This avoids import cycles between services and services/providers
type CursorServiceInterface interface {
	ChatCompletion(ctx context.Context, request *models.ChatCompletionRequest) (<-chan interface{}, *middleware.CursorSessionInfo, error)
}

// CursorProvider wraps the existing CursorService as a ProviderClient
type CursorProvider struct {
	cursorService CursorServiceInterface
}

// NewCursorProvider creates a new Cursor provider instance
func NewCursorProvider(cursorService CursorServiceInterface) *CursorProvider {
	return &CursorProvider{
		cursorService: cursorService,
	}
}

// IsAvailable returns true if the provider is properly configured
// Cursor provider is available only if cursorService is configured
func (p *CursorProvider) IsAvailable() bool {
	return p.cursorService != nil
}

// GetProviderName returns the provider identifier
func (p *CursorProvider) GetProviderName() string {
	return "cursor"
}

// GetSupportedModels returns the list of models supported by this provider
// Returns all Cursor-supported models
func (p *CursorProvider) GetSupportedModels() []models.ModelInfo {
	// Cursor supports a wide range of models as fallback
	return []models.ModelInfo{
		// Claude models
		{
			ID:            "claude-3.5-sonnet",
			Name:          "Claude 3.5 Sonnet",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-3.5-haiku",
			Name:          "Claude 3.5 Haiku",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    0.80,
			OutputPrice:   4.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-3.7-sonnet",
			Name:          "Claude 3.7 Sonnet",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4-sonnet",
			Name:          "Claude 4 Sonnet",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4.5-sonnet",
			Name:          "Claude 4.5 Sonnet",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    3.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4-opus",
			Name:          "Claude 4 Opus",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   75.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4.1-opus",
			Name:          "Claude 4.1 Opus",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   75.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4.5-opus",
			Name:          "Claude 4.5 Opus",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    15.00,
			OutputPrice:   75.00,
			IsAvailable:   true,
		},
		{
			ID:            "claude-4.5-haiku",
			Name:          "Claude 4.5 Haiku",
			Provider:      "cursor",
			ContextWindow: 200000,
			InputPrice:    0.80,
			OutputPrice:   4.00,
			IsAvailable:   true,
		},
		// GPT models
		{
			ID:            "gpt-4o",
			Name:          "OpenAI GPT-4o",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    2.50,
			OutputPrice:   10.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5.2",
			Name:          "GPT-5.2",
			Provider:      "cursor",
			ContextWindow: 512000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5",
			Name:          "GPT-5",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5.1",
			Name:          "GPT-5.1",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5-codex",
			Name:          "GPT-5 Codex",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5.1-codex",
			Name:          "GPT-5.1 Codex",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5.1-codex-max",
			Name:          "GPT-5.1 Codex Max",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    10.00,
			OutputPrice:   30.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5-mini",
			Name:          "GPT-5 Mini",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    1.00,
			OutputPrice:   3.00,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-5-nano",
			Name:          "GPT-5 Nano",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    0.50,
			OutputPrice:   1.50,
			IsAvailable:   true,
		},
		{
			ID:            "gpt-4.1",
			Name:          "GPT-4.1",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    10.00,
			OutputPrice:   30.00,
			IsAvailable:   true,
		},
		// O series models
		{
			ID:            "o3",
			Name:          "OpenAI O-Series",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    15.00,
			OutputPrice:   60.00,
			IsAvailable:   true,
		},
		{
			ID:            "o4-mini",
			Name:          "O4 Mini",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    3.00,
			OutputPrice:   12.00,
			IsAvailable:   true,
		},
		// Gemini models
		{
			ID:            "gemini-2.5-pro",
			Name:          "Gemini 2.5 Pro",
			Provider:      "cursor",
			ContextWindow: 1000000,
			InputPrice:    1.25,
			OutputPrice:   5.00,
			IsAvailable:   true,
		},
		{
			ID:            "gemini-2.5-flash",
			Name:          "Gemini 2.5 Flash",
			Provider:      "cursor",
			ContextWindow: 1000000,
			InputPrice:    0.075,
			OutputPrice:   0.30,
			IsAvailable:   true,
		},
		{
			ID:            "gemini-3-pro-preview",
			Name:          "Gemini 3 Pro Preview",
			Provider:      "cursor",
			ContextWindow: 1000000,
			InputPrice:    1.25,
			OutputPrice:   5.00,
			IsAvailable:   true,
		},
		// DeepSeek models
		{
			ID:            "deepseek-r1",
			Name:          "DeepSeek R1",
			Provider:      "cursor",
			ContextWindow: 64000,
			InputPrice:    0.55,
			OutputPrice:   2.19,
			IsAvailable:   true,
		},
		{
			ID:            "deepseek-v3.1",
			Name:          "DeepSeek V3.1",
			Provider:      "cursor",
			ContextWindow: 64000,
			InputPrice:    0.27,
			OutputPrice:   1.10,
			IsAvailable:   true,
		},
		// Other models
		{
			ID:            "kimi-k2-instruct",
			Name:          "Kimi K2 Instruct",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    2.00,
			OutputPrice:   8.00,
			IsAvailable:   true,
		},
		{
			ID:            "grok-3",
			Name:          "Grok 3",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "grok-3-mini",
			Name:          "Grok 3 Mini",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    2.00,
			OutputPrice:   8.00,
			IsAvailable:   true,
		},
		{
			ID:            "grok-4",
			Name:          "Grok 4",
			Provider:      "cursor",
			ContextWindow: 128000,
			InputPrice:    5.00,
			OutputPrice:   15.00,
			IsAvailable:   true,
		},
		{
			ID:            "code-supernova-1-million",
			Name:          "Code Supernova 1M",
			Provider:      "cursor",
			ContextWindow: 1000000,
			InputPrice:    10.00,
			OutputPrice:   40.00,
			IsAvailable:   true,
		},
	}
}

// ChatCompletion sends a chat request and returns a streaming channel
// Converts ChatRequest to CursorService format and converts streaming response to unified format
func (p *CursorProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error) {
	// Check if cursorService is available
	if p.cursorService == nil {
		return nil, fmt.Errorf("cursor service not initialized")
	}
	
	// Convert ChatRequest to CursorService format
	cursorReq := &models.ChatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   req.Stream,
	}

	if req.MaxTokens > 0 {
		maxTokens := req.MaxTokens
		cursorReq.MaxTokens = &maxTokens
	}

	if req.Temperature > 0 {
		temperature := req.Temperature
		cursorReq.Temperature = &temperature
	}

	// Call existing CursorService
	cursorStreamChan, _, err := p.cursorService.ChatCompletion(ctx, cursorReq)
	if err != nil {
		return nil, fmt.Errorf("cursor service error: %w", err)
	}

	// Create channel for unified StreamEvent format
	eventChan := make(chan models.StreamEvent)

	// Start goroutine to convert Cursor streaming format to unified format
	go p.convertCursorStream(cursorStreamChan, eventChan)

	return eventChan, nil
}

// convertCursorStream converts Cursor's streaming format to unified StreamEvent format
func (p *CursorProvider) convertCursorStream(cursorChan <-chan interface{}, eventChan chan<- models.StreamEvent) {
	defer close(eventChan)

	var totalUsage *models.TokenUsage
	hasStarted := false

	for event := range cursorChan {
		// Send start event on first message
		if !hasStarted {
			eventChan <- models.StreamEvent{
				Type: "start",
			}
			hasStarted = true
		}

		// Handle different event types from Cursor
		switch v := event.(type) {
		case string:
			// String events are typically JSON-encoded data
			var cursorEvent models.CursorEventData
			if err := json.Unmarshal([]byte(v), &cursorEvent); err != nil {
				// If not JSON, treat as plain text content
				eventChan <- models.StreamEvent{
					Type:    "content",
					Content: v,
				}
				continue
			}

			// Process based on Cursor event type
			switch cursorEvent.Type {
			case "delta":
				if cursorEvent.Delta != "" {
					eventChan <- models.StreamEvent{
						Type:    "content",
						Content: cursorEvent.Delta,
					}
				}

			case "error":
				eventChan <- models.StreamEvent{
					Type:  "error",
					Error: cursorEvent.ErrorText,
				}
				return

			case "done":
				// Extract usage information if available
				if cursorEvent.MessageMetadata != nil && cursorEvent.MessageMetadata.Usage != nil {
					totalUsage = &models.TokenUsage{
						PromptTokens:     cursorEvent.MessageMetadata.Usage.InputTokens,
						CompletionTokens: cursorEvent.MessageMetadata.Usage.OutputTokens,
						TotalTokens:      cursorEvent.MessageMetadata.Usage.TotalTokens,
					}
				}
			}

		case models.CursorEventData:
			// Direct CursorEventData struct
			switch v.Type {
			case "delta":
				if v.Delta != "" {
					eventChan <- models.StreamEvent{
						Type:    "content",
						Content: v.Delta,
					}
				}

			case "error":
				eventChan <- models.StreamEvent{
					Type:  "error",
					Error: v.ErrorText,
				}
				return

			case "done":
				// Extract usage information if available
				if v.MessageMetadata != nil && v.MessageMetadata.Usage != nil {
					totalUsage = &models.TokenUsage{
						PromptTokens:     v.MessageMetadata.Usage.InputTokens,
						CompletionTokens: v.MessageMetadata.Usage.OutputTokens,
						TotalTokens:      v.MessageMetadata.Usage.TotalTokens,
					}
				}
			}

		case error:
			// Error from Cursor service
			eventChan <- models.StreamEvent{
				Type:  "error",
				Error: v.Error(),
			}
			return

		default:
			// Try to handle as JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				continue
			}

			var cursorEvent models.CursorEventData
			if err := json.Unmarshal(jsonBytes, &cursorEvent); err != nil {
				continue
			}

			// Process the event
			if cursorEvent.Type == "delta" && cursorEvent.Delta != "" {
				eventChan <- models.StreamEvent{
					Type:    "content",
					Content: cursorEvent.Delta,
				}
			} else if cursorEvent.Type == "error" {
				eventChan <- models.StreamEvent{
					Type:  "error",
					Error: cursorEvent.ErrorText,
				}
				return
			}
		}
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
