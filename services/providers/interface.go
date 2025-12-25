package providers

import (
	"context"

	"Curry2API-go/models"
)

// ProviderClient defines the interface for AI provider implementations
type ProviderClient interface {
	// ChatCompletion sends a chat request and returns a streaming channel
	ChatCompletion(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamEvent, error)

	// GetSupportedModels returns the list of models supported by this provider
	GetSupportedModels() []models.ModelInfo

	// GetProviderName returns the provider identifier
	GetProviderName() string

	// IsAvailable returns true if the provider is properly configured
	IsAvailable() bool
}
