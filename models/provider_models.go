package models

// StreamEvent represents a unified streaming event from any provider
type StreamEvent struct {
	Type    string      `json:"type"`              // "start", "content", "usage", "done", "error"
	Content string      `json:"content,omitempty"` // Text content for "content" type events
	Tokens  *TokenUsage `json:"tokens,omitempty"`  // Token usage for "usage" type events
	Error   string      `json:"error,omitempty"`   // Error message for "error" type events
}

// TokenUsage represents token consumption information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo represents information about an AI model
type ModelInfo struct {
	ID            string  `json:"id"`              // Model identifier (e.g., "gpt-4o")
	Name          string  `json:"name"`            // Human-readable name
	Provider      string  `json:"provider"`        // Provider name (e.g., "openai")
	ContextWindow int     `json:"context_window"`  // Maximum context length in tokens
	InputPrice    float64 `json:"input_price"`     // Price per 1M input tokens
	OutputPrice   float64 `json:"output_price"`    // Price per 1M output tokens
	IsAvailable   bool    `json:"is_available"`    // Whether the provider is configured
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}
