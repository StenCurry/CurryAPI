package services

import (
	"encoding/json"
	"fmt"

	"Curry2API-go/models"
)

// StreamConverter provides functions to convert provider-specific stream events
// to the unified StreamEvent format.
// This centralizes all stream format conversion logic for consistency.

// OpenAIStreamChunk represents a chunk from OpenAI's streaming response
type OpenAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// AnthropicStreamChunk represents a chunk from Anthropic's streaming response
type AnthropicStreamChunk struct {
	Type    string `json:"type"`
	Message *struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Role  string `json:"role"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Index        int `json:"index,omitempty"`
	ContentBlock *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type       string  `json:"type"`
		Text       string  `json:"text,omitempty"`
		StopReason *string `json:"stop_reason,omitempty"`
	} `json:"delta,omitempty"`
	Usage *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// GoogleStreamChunk represents a chunk from Google's streaming response
type GoogleStreamChunk struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"candidates,omitempty"`
	UsageMetadata *struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata,omitempty"`
}

// DeepSeekStreamChunk represents a chunk from DeepSeek's streaming response
// DeepSeek uses OpenAI-compatible format
type DeepSeekStreamChunk = OpenAIStreamChunk

// ConvertOpenAIStream converts an OpenAI stream chunk to unified StreamEvent
// Returns nil if the chunk doesn't produce a meaningful event
func ConvertOpenAIStream(data []byte) (*models.StreamEvent, error) {
	var chunk OpenAIStreamChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI stream chunk: %w", err)
	}

	// Check for usage information (typically sent at the end)
	if chunk.Usage != nil {
		return &models.StreamEvent{
			Type: "usage",
			Tokens: &models.TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			},
		}, nil
	}

	// Process choices
	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]

		// Check for content delta
		if choice.Delta.Content != "" {
			return &models.StreamEvent{
				Type:    "content",
				Content: choice.Delta.Content,
			}, nil
		}

		// Check for finish reason (indicates completion)
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			return &models.StreamEvent{
				Type: "done",
			}, nil
		}
	}

	// No meaningful event to produce
	return nil, nil
}

// ConvertAnthropicStream converts an Anthropic stream chunk to unified StreamEvent
// eventType is the SSE event type (e.g., "message_start", "content_block_delta")
// Returns nil if the chunk doesn't produce a meaningful event
func ConvertAnthropicStream(eventType string, data []byte) (*models.StreamEvent, *models.TokenUsage, error) {
	var chunk AnthropicStreamChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Anthropic stream chunk: %w", err)
	}

	var tokenUsage *models.TokenUsage

	switch eventType {
	case "message_start":
		// Extract initial token usage (input tokens)
		if chunk.Message != nil && chunk.Message.Usage != nil {
			tokenUsage = &models.TokenUsage{
				PromptTokens: chunk.Message.Usage.InputTokens,
			}
		}
		return &models.StreamEvent{
			Type: "start",
		}, tokenUsage, nil

	case "content_block_delta":
		// Send content delta
		if chunk.Delta != nil && chunk.Delta.Text != "" {
			return &models.StreamEvent{
				Type:    "content",
				Content: chunk.Delta.Text,
			}, nil, nil
		}

	case "message_delta":
		// Extract output token usage
		if chunk.Usage != nil {
			tokenUsage = &models.TokenUsage{
				CompletionTokens: chunk.Usage.OutputTokens,
			}
		}
		return nil, tokenUsage, nil

	case "message_stop":
		return &models.StreamEvent{
			Type: "done",
		}, nil, nil

	case "error":
		return &models.StreamEvent{
			Type:  "error",
			Error: "Anthropic API error",
		}, nil, nil
	}

	// No meaningful event to produce
	return nil, nil, nil
}

// ConvertGoogleStream converts a Google stream chunk to unified StreamEvent
// Returns nil if the chunk doesn't produce a meaningful event
func ConvertGoogleStream(data []byte) (*models.StreamEvent, *models.TokenUsage, error) {
	var chunk GoogleStreamChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Google stream chunk: %w", err)
	}

	var tokenUsage *models.TokenUsage

	// Extract usage metadata if present
	if chunk.UsageMetadata != nil {
		tokenUsage = &models.TokenUsage{
			PromptTokens:     chunk.UsageMetadata.PromptTokenCount,
			CompletionTokens: chunk.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      chunk.UsageMetadata.TotalTokenCount,
		}
	}

	// Process candidates
	if len(chunk.Candidates) > 0 {
		candidate := chunk.Candidates[0]

		// Extract content from parts
		var content string
		if len(candidate.Content.Parts) > 0 {
			for _, part := range candidate.Content.Parts {
				content += part.Text
			}
		}

		if content != "" {
			return &models.StreamEvent{
				Type:    "content",
				Content: content,
			}, tokenUsage, nil
		}

		// Check for finish reason
		if candidate.FinishReason != "" {
			return &models.StreamEvent{
				Type: "done",
			}, tokenUsage, nil
		}
	}

	// Return usage if we have it but no content
	if tokenUsage != nil {
		return &models.StreamEvent{
			Type:   "usage",
			Tokens: tokenUsage,
		}, nil, nil
	}

	// No meaningful event to produce
	return nil, nil, nil
}

// ConvertDeepSeekStream converts a DeepSeek stream chunk to unified StreamEvent
// DeepSeek uses OpenAI-compatible format, so this delegates to ConvertOpenAIStream
func ConvertDeepSeekStream(data []byte) (*models.StreamEvent, error) {
	return ConvertOpenAIStream(data)
}

// ValidStreamEventTypes returns the list of valid StreamEvent types
func ValidStreamEventTypes() []string {
	return []string{"start", "content", "usage", "done", "error"}
}

// IsValidStreamEventType checks if a type is a valid StreamEvent type
func IsValidStreamEventType(eventType string) bool {
	for _, t := range ValidStreamEventTypes() {
		if t == eventType {
			return true
		}
	}
	return false
}

// CreateStartEvent creates a start StreamEvent
func CreateStartEvent() models.StreamEvent {
	return models.StreamEvent{
		Type: "start",
	}
}

// CreateContentEvent creates a content StreamEvent
func CreateContentEvent(content string) models.StreamEvent {
	return models.StreamEvent{
		Type:    "content",
		Content: content,
	}
}

// CreateUsageEvent creates a usage StreamEvent
func CreateUsageEvent(promptTokens, completionTokens int) models.StreamEvent {
	return models.StreamEvent{
		Type: "usage",
		Tokens: &models.TokenUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// CreateDoneEvent creates a done StreamEvent
func CreateDoneEvent() models.StreamEvent {
	return models.StreamEvent{
		Type: "done",
	}
}

// CreateErrorEvent creates an error StreamEvent
func CreateErrorEvent(errorMsg string) models.StreamEvent {
	return models.StreamEvent{
		Type:  "error",
		Error: errorMsg,
	}
}
