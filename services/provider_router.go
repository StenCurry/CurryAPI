package services

import (
	"Curry2API-go/config"
	"Curry2API-go/models"
	"Curry2API-go/services/providers"
	"fmt"
	"strings"
)

// ProviderRouter routes model requests to the appropriate provider
type ProviderRouter struct {
	providers map[string]providers.ProviderClient
	config    *config.Config
}

// NewProviderRouter creates a new provider router with the given configuration
func NewProviderRouter(cfg *config.Config) *ProviderRouter {
	router := &ProviderRouter{
		providers: make(map[string]providers.ProviderClient),
		config:    cfg,
	}
	
	// Initialize providers based on available API keys
	
	// Initialize OpenAI provider if API key is configured
	if cfg.Providers.OpenAI.APIKey != "" {
		openaiProvider := providers.NewOpenAIProvider(
			cfg.Providers.OpenAI.APIKey,
			cfg.Providers.OpenAI.BaseURL,
		)
		router.providers["openai"] = openaiProvider
	}
	
	// Initialize Anthropic provider if API key is configured
	if cfg.Providers.Anthropic.APIKey != "" {
		anthropicProvider := providers.NewAnthropicProvider(
			cfg.Providers.Anthropic.APIKey,
			cfg.Providers.Anthropic.BaseURL,
		)
		router.providers["anthropic"] = anthropicProvider
	}
	
	// Initialize Google provider if API key is configured
	if cfg.Providers.Google.APIKey != "" {
		googleProvider := providers.NewGoogleProvider(
			cfg.Providers.Google.APIKey,
		)
		router.providers["google"] = googleProvider
	}
	
	// Initialize DeepSeek provider if API key is configured
	if cfg.Providers.DeepSeek.APIKey != "" {
		deepseekProvider := providers.NewDeepSeekProvider(
			cfg.Providers.DeepSeek.APIKey,
			cfg.Providers.DeepSeek.BaseURL,
		)
		router.providers["deepseek"] = deepseekProvider
	}
	
	return router
}

// GetProvider returns the appropriate provider for the given model
// Always uses Cursor provider as the primary provider for all models
// This ensures consistent behavior using the CursorSession system
func (r *ProviderRouter) GetProvider(model string) (providers.ProviderClient, error) {
	// Always use Cursor provider as the primary provider
	// Cursor provider supports all models through the CursorSession system
	if cursorProvider, exists := r.providers["cursor"]; exists && cursorProvider.IsAvailable() {
		return cursorProvider, nil
	}
	
	// If Cursor is not available, try to find an alternative provider based on model
	modelLower := strings.ToLower(model)
	
	// Helper function to get provider
	getProvider := func(providerName string) (providers.ProviderClient, error) {
		if provider, exists := r.providers[providerName]; exists && provider.IsAvailable() {
			return provider, nil
		}
		return nil, fmt.Errorf("PROVIDER_NOT_AVAILABLE: %s provider is not available", providerName)
	}
	
	// Route based on model name prefix as fallback
	// OpenAI models: gpt-*, o1*, o3*, o4*
	if strings.HasPrefix(modelLower, "gpt-") || 
	   strings.HasPrefix(modelLower, "o1") || 
	   strings.HasPrefix(modelLower, "o3") ||
	   strings.HasPrefix(modelLower, "o4") {
		return getProvider("openai")
	}
	
	// Anthropic models: claude-*
	if strings.HasPrefix(modelLower, "claude-") {
		return getProvider("anthropic")
	}
	
	// Google models: gemini-*
	if strings.HasPrefix(modelLower, "gemini-") {
		return getProvider("google")
	}
	
	// DeepSeek models: deepseek-*
	if strings.HasPrefix(modelLower, "deepseek-") {
		return getProvider("deepseek")
	}
	
	return nil, fmt.Errorf("PROVIDER_NOT_AVAILABLE: No provider available for model %s", model)
}

// GetAvailableProviders returns list of configured providers
func (r *ProviderRouter) GetAvailableProviders() []string {
	available := make([]string, 0, len(r.providers))
	for name, provider := range r.providers {
		if provider.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}

// GetAllModels returns all available models from all providers
func (r *ProviderRouter) GetAllModels() []models.ModelInfo {
	allModels := make([]models.ModelInfo, 0)
	
	for _, provider := range r.providers {
		models := provider.GetSupportedModels()
		allModels = append(allModels, models...)
	}
	
	// 添加 OpenRouter 免费模型
	openRouterModels := GetOpenRouterFreeModelInfos()
	allModels = append(allModels, openRouterModels...)
	
	return allModels
}

// RegisterProvider registers a provider with the router
// This is used for testing and for adding providers after initialization
func (r *ProviderRouter) RegisterProvider(name string, provider providers.ProviderClient) {
	r.providers[name] = provider
}

// NewCursorProvider creates a new Cursor provider instance
// This is a wrapper function to avoid exposing the providers package directly
func NewCursorProvider(cursorService providers.CursorServiceInterface) providers.ProviderClient {
	return providers.NewCursorProvider(cursorService)
}
