package utils

import (
	"Curry2API-go/models"
	"strings"
)

// ExtractTokenUsage extracts token counts from Cursor API response
func ExtractTokenUsage(eventData models.CursorEventData) *models.Usage {
	if eventData.MessageMetadata != nil && eventData.MessageMetadata.Usage != nil {
		return &models.Usage{
			PromptTokens:     eventData.MessageMetadata.Usage.InputTokens,
			CompletionTokens: eventData.MessageMetadata.Usage.OutputTokens,
			TotalTokens:      eventData.MessageMetadata.Usage.TotalTokens,
		}
	}
	return nil
}

// EstimateTokenUsage estimates tokens based on message content
// Uses rough approximation: 1 token ≈ 4 characters or 0.75 words
func EstimateTokenUsage(messages []models.Message) int {
	totalChars := 0
	for _, msg := range messages {
		// Handle Content which can be string or interface{}
		if content, ok := msg.Content.(string); ok {
			totalChars += len(content)
		}
	}
	// Rough estimation: 4 characters per token
	estimatedTokens := totalChars / 4
	
	// Add minimum baseline
	if estimatedTokens < 10 {
		estimatedTokens = 10
	}
	
	return estimatedTokens
}

// EstimateTokensFromText estimates tokens from a single text string
func EstimateTokensFromText(text string) int {
	if text == "" {
		return 0
	}
	
	// Count characters
	charCount := len(text)
	
	// Rough estimation: 4 characters per token
	tokens := charCount / 4
	
	// Minimum 1 token for non-empty text
	if tokens < 1 {
		tokens = 1
	}
	
	return tokens
}

// EstimateResponseTokens estimates tokens needed for a response based on request
// This is a heuristic that can be tuned based on actual usage patterns
func EstimateResponseTokens(requestTokens int, multiplier float64) int {
	// Typical response is 2-3x the request size for conversational AI
	if multiplier <= 0 {
		multiplier = 2.5
	}
	
	estimated := int(float64(requestTokens) * multiplier)
	
	// Cap at reasonable maximum
	maxTokens := 8000
	if estimated > maxTokens {
		estimated = maxTokens
	}
	
	// Minimum response size
	if estimated < 100 {
		estimated = 100
	}
	
	return estimated
}

// EstimateTotalRequestTokens estimates total tokens for a complete request/response cycle
func EstimateTotalRequestTokens(messages []models.Message, estimationMultiplier float64) int {
	requestTokens := EstimateTokenUsage(messages)
	responseTokens := EstimateResponseTokens(requestTokens, estimationMultiplier)
	return requestTokens + responseTokens
}

// CalculateTokensFromContent calculates tokens more accurately using word count
func CalculateTokensFromContent(content string) int {
	if content == "" {
		return 0
	}
	
	// Trim whitespace
	content = strings.TrimSpace(content)
	
	// Count words (split by whitespace)
	words := strings.Fields(content)
	wordCount := len(words)
	
	// Average: 1 token ≈ 0.75 words, so tokens = words / 0.75 = words * 1.33
	tokens := int(float64(wordCount) * 1.33)
	
	// Also consider character count for languages without spaces (e.g., Chinese)
	charTokens := len(content) / 4
	
	// Use the larger estimate
	if charTokens > tokens {
		tokens = charTokens
	}
	
	// Minimum 1 token for non-empty content
	if tokens < 1 {
		tokens = 1
	}
	
	return tokens
}
