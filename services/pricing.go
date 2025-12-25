package services

import (
	"strings"
)

// ModelPricing represents pricing information for a model
type ModelPricing struct {
	Model       string  `json:"model"`
	Provider    string  `json:"provider"`
	InputPrice  float64 `json:"input_price"`  // Price per 1M input tokens
	OutputPrice float64 `json:"output_price"` // Price per 1M output tokens
}

// pricingTable contains pricing information for all supported models
// Prices are in USD per 1M tokens
var pricingTable = map[string]ModelPricing{
	// OpenAI models
	"gpt-5.2": {
		Model:       "gpt-5.2",
		Provider:    "openai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"gpt-5": {
		Model:       "gpt-5",
		Provider:    "openai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"gpt-5.1": {
		Model:       "gpt-5.1",
		Provider:    "openai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"gpt-5-codex": {
		Model:       "gpt-5-codex",
		Provider:    "openai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"gpt-5.1-codex": {
		Model:       "gpt-5.1-codex",
		Provider:    "openai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"gpt-5.1-codex-max": {
		Model:       "gpt-5.1-codex-max",
		Provider:    "openai",
		InputPrice:  10.00,
		OutputPrice: 30.00,
	},
	"gpt-5-mini": {
		Model:       "gpt-5-mini",
		Provider:    "openai",
		InputPrice:  1.00,
		OutputPrice: 3.00,
	},
	"gpt-5-nano": {
		Model:       "gpt-5-nano",
		Provider:    "openai",
		InputPrice:  0.50,
		OutputPrice: 1.50,
	},
	"gpt-4.1": {
		Model:       "gpt-4.1",
		Provider:    "openai",
		InputPrice:  10.00,
		OutputPrice: 30.00,
	},
	"gpt-4o": {
		Model:       "gpt-4o",
		Provider:    "openai",
		InputPrice:  2.50,
		OutputPrice: 10.00,
	},
	"gpt-4o-mini": {
		Model:       "gpt-4o-mini",
		Provider:    "openai",
		InputPrice:  0.15,
		OutputPrice: 0.60,
	},
	"gpt-4-turbo": {
		Model:       "gpt-4-turbo",
		Provider:    "openai",
		InputPrice:  10.00,
		OutputPrice: 30.00,
	},
	"gpt-4": {
		Model:       "gpt-4",
		Provider:    "openai",
		InputPrice:  30.00,
		OutputPrice: 60.00,
	},
	"gpt-3.5-turbo": {
		Model:       "gpt-3.5-turbo",
		Provider:    "openai",
		InputPrice:  0.50,
		OutputPrice: 1.50,
	},
	"o1": {
		Model:       "o1",
		Provider:    "openai",
		InputPrice:  15.00,
		OutputPrice: 60.00,
	},
	"o1-mini": {
		Model:       "o1-mini",
		Provider:    "openai",
		InputPrice:  3.00,
		OutputPrice: 12.00,
	},
	"o3": {
		Model:       "o3",
		Provider:    "openai",
		InputPrice:  15.00,
		OutputPrice: 60.00,
	},
	"o3-mini": {
		Model:       "o3-mini",
		Provider:    "openai",
		InputPrice:  3.00,
		OutputPrice: 12.00,
	},

	// Anthropic models (legacy names)
	"claude-3-5-sonnet-20241022": {
		Model:       "claude-3-5-sonnet-20241022",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-3-5-haiku-20241022": {
		Model:       "claude-3-5-haiku-20241022",
		Provider:    "anthropic",
		InputPrice:  0.80,
		OutputPrice: 4.00,
	},
	"claude-3-opus-20240229": {
		Model:       "claude-3-opus-20240229",
		Provider:    "anthropic",
		InputPrice:  15.00,
		OutputPrice: 75.00,
	},
	"claude-3-sonnet-20240229": {
		Model:       "claude-3-sonnet-20240229",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-3-haiku-20240307": {
		Model:       "claude-3-haiku-20240307",
		Provider:    "anthropic",
		InputPrice:  0.25,
		OutputPrice: 1.25,
	},
	// Anthropic models (normalized names)
	"claude-3.5-sonnet": {
		Model:       "claude-3.5-sonnet",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-3.5-haiku": {
		Model:       "claude-3.5-haiku",
		Provider:    "anthropic",
		InputPrice:  0.80,
		OutputPrice: 4.00,
	},
	"claude-3.7-sonnet": {
		Model:       "claude-3.7-sonnet",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-4-sonnet": {
		Model:       "claude-4-sonnet",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-4.5-sonnet": {
		Model:       "claude-4.5-sonnet",
		Provider:    "anthropic",
		InputPrice:  3.00,
		OutputPrice: 15.00,
	},
	"claude-4-opus": {
		Model:       "claude-4-opus",
		Provider:    "anthropic",
		InputPrice:  15.00,
		OutputPrice: 75.00,
	},
	"claude-4.1-opus": {
		Model:       "claude-4.1-opus",
		Provider:    "anthropic",
		InputPrice:  15.00,
		OutputPrice: 75.00,
	},
	"claude-4.5-opus": {
		Model:       "claude-4.5-opus",
		Provider:    "anthropic",
		InputPrice:  15.00,
		OutputPrice: 75.00,
	},
	"claude-4.5-haiku": {
		Model:       "claude-4.5-haiku",
		Provider:    "anthropic",
		InputPrice:  0.80,
		OutputPrice: 4.00,
	},

	// Google models (legacy)
	"gemini-1.5-pro": {
		Model:       "gemini-1.5-pro",
		Provider:    "google",
		InputPrice:  1.25,
		OutputPrice: 5.00,
	},
	"gemini-1.5-flash": {
		Model:       "gemini-1.5-flash",
		Provider:    "google",
		InputPrice:  0.075,
		OutputPrice: 0.30,
	},
	"gemini-pro": {
		Model:       "gemini-pro",
		Provider:    "google",
		InputPrice:  0.50,
		OutputPrice: 1.50,
	},
	// Google models (new)
	"gemini-2.5-pro": {
		Model:       "gemini-2.5-pro",
		Provider:    "google",
		InputPrice:  1.25,
		OutputPrice: 5.00,
	},
	"gemini-2.5-flash": {
		Model:       "gemini-2.5-flash",
		Provider:    "google",
		InputPrice:  0.075,
		OutputPrice: 0.30,
	},
	"gemini-3-pro-preview": {
		Model:       "gemini-3-pro-preview",
		Provider:    "google",
		InputPrice:  1.25,
		OutputPrice: 5.00,
	},

	// DeepSeek models (legacy)
	"deepseek-chat": {
		Model:       "deepseek-chat",
		Provider:    "deepseek",
		InputPrice:  0.14,
		OutputPrice: 0.28,
	},
	"deepseek-coder": {
		Model:       "deepseek-coder",
		Provider:    "deepseek",
		InputPrice:  0.14,
		OutputPrice: 0.28,
	},
	"deepseek-reasoner": {
		Model:       "deepseek-reasoner",
		Provider:    "deepseek",
		InputPrice:  0.55,
		OutputPrice: 2.19,
	},
	// DeepSeek models (new)
	"deepseek-r1": {
		Model:       "deepseek-r1",
		Provider:    "deepseek",
		InputPrice:  0.55,
		OutputPrice: 2.19,
	},
	"deepseek-v3.1": {
		Model:       "deepseek-v3.1",
		Provider:    "deepseek",
		InputPrice:  0.27,
		OutputPrice: 1.10,
	},

	// O-series models
	"o4-mini": {
		Model:       "o4-mini",
		Provider:    "openai",
		InputPrice:  3.00,
		OutputPrice: 12.00,
	},

	// Kimi models
	"kimi-k2-instruct": {
		Model:       "kimi-k2-instruct",
		Provider:    "moonshot",
		InputPrice:  2.00,
		OutputPrice: 8.00,
	},

	// Grok models
	"grok-3": {
		Model:       "grok-3",
		Provider:    "xai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},
	"grok-3-mini": {
		Model:       "grok-3-mini",
		Provider:    "xai",
		InputPrice:  2.00,
		OutputPrice: 8.00,
	},
	"grok-4": {
		Model:       "grok-4",
		Provider:    "xai",
		InputPrice:  5.00,
		OutputPrice: 15.00,
	},

	// Code Supernova
	"code-supernova-1-million": {
		Model:       "code-supernova-1-million",
		Provider:    "supernova",
		InputPrice:  10.00,
		OutputPrice: 40.00,
	},
}

// GetModelPricing returns the pricing information for a given model
// Returns nil if the model is not found in the pricing table
func GetModelPricing(model string) *ModelPricing {
	modelLower := strings.ToLower(model)
	if pricing, exists := pricingTable[modelLower]; exists {
		return &pricing
	}
	return nil
}

// CalculateCost calculates the cost for a given model and token usage
// Returns the cost in USD
// Formula: (prompt_tokens * input_price + completion_tokens * output_price) / 1,000,000
func CalculateCost(model string, promptTokens, completionTokens int) float64 {
	pricing := GetModelPricing(model)
	if pricing == nil {
		return 0.0
	}
	return CalculateCostWithPricing(promptTokens, completionTokens, pricing.InputPrice, pricing.OutputPrice)
}

// CalculateCostWithPricing calculates the cost given token counts and prices directly
// This is useful for testing and when pricing is already known
// Formula: (prompt_tokens * input_price + completion_tokens * output_price) / 1,000,000
func CalculateCostWithPricing(promptTokens, completionTokens int, inputPrice, outputPrice float64) float64 {
	inputCost := float64(promptTokens) * inputPrice
	outputCost := float64(completionTokens) * outputPrice
	return (inputCost + outputCost) / 1_000_000
}

// GetAllPricing returns all pricing information
func GetAllPricing() map[string]ModelPricing {
	// Return a copy to prevent modification
	result := make(map[string]ModelPricing, len(pricingTable))
	for k, v := range pricingTable {
		result[k] = v
	}
	return result
}


// GetProviderFromModel determines the provider name from a model name
// This is used for logging and usage tracking
func GetProviderFromModel(model string) string {
	modelLower := strings.ToLower(model)

	// OpenAI models: gpt-*, o1*, o3*, o4*
	if strings.HasPrefix(modelLower, "gpt-") ||
		strings.HasPrefix(modelLower, "o1") ||
		strings.HasPrefix(modelLower, "o3") ||
		strings.HasPrefix(modelLower, "o4") {
		return "openai"
	}

	// Anthropic models: claude-*
	if strings.HasPrefix(modelLower, "claude-") {
		return "anthropic"
	}

	// Google models: gemini-*
	if strings.HasPrefix(modelLower, "gemini-") {
		return "google"
	}

	// DeepSeek models: deepseek-*
	if strings.HasPrefix(modelLower, "deepseek-") {
		return "deepseek"
	}

	// Default to cursor for unknown models
	return "cursor"
}
