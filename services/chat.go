package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"Curry2API-go/config"
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/services/providers"

	"github.com/sirupsen/logrus"
)

// Billing-related errors
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// Chat service errors
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrUnauthorized         = errors.New("unauthorized access to conversation")
	ErrEmptyMessage         = errors.New("message content cannot be empty")
	ErrAIServiceUnavailable = errors.New("AI service temporarily unavailable")
	ErrAIServiceTimeout     = errors.New("AI service request timeout")
	ErrInvalidModel         = errors.New("invalid model specified")
)

// Provider-specific errors are defined in provider_errors.go
// ErrProviderNotAvailable, ErrInvalidAPIKey, ErrRateLimited, ErrProviderError, ErrTimeout, ErrContextTooLong

// SendMessageRequest represents a request to send a message in a conversation
type SendMessageRequest struct {
	ConversationID int64
	UserID         int64
	Content        string
	Model          string // Optional: override conversation model
}

// SendMessageResponse represents the response from sending a message
type SendMessageResponse struct {
	UserMessage *models.ChatMessage
	StreamChan  <-chan models.StreamEvent
}

// ChatService handles chat business logic including message processing and AI integration
// Requirements: 2.1-2.6, 10.1-10.5
type ChatService struct {
	cursorService  *CursorService
	providerRouter *ProviderRouter
	config         *config.Config
}

// NewChatService creates a new ChatService instance
// Updated to accept ProviderRouter for multi-provider support
func NewChatService(cursorService *CursorService, cfg *config.Config) *ChatService {
	return &ChatService{
		cursorService: cursorService,
		config:        cfg,
	}
}

// NewChatServiceWithRouter creates a new ChatService instance with ProviderRouter
// Requirements: 2.1-2.6
func NewChatServiceWithRouter(cursorService *CursorService, providerRouter *ProviderRouter, cfg *config.Config) *ChatService {
	return &ChatService{
		cursorService:  cursorService,
		providerRouter: providerRouter,
		config:         cfg,
	}
}

// SetProviderRouter sets the provider router for the chat service
func (s *ChatService) SetProviderRouter(router *ProviderRouter) {
	s.providerRouter = router
}

// mapProviderError maps provider-specific errors to user-friendly errors
// Requirements: 10.1-10.5, 10.6
func mapProviderError(err error, provider string, model string, requestID string) error {
	if err == nil {
		return nil
	}

	// Use centralized error wrapping and logging
	providerErr := WrapError(err, provider, model, requestID)
	
	// Log the error with structured fields (Requirements: 10.6)
	LogProviderError(providerErr)

	return providerErr
}

// BuildContext retrieves all messages from a conversation to build context for AI requests
// Requirements: 2.3 - Include all previous messages in the conversation as context
// **Feature: online-chat, Property 7: Context Building**
// **Validates: Requirements 2.3**
func (s *ChatService) BuildContext(conversationID int64) ([]models.Message, error) {
	// Get all messages from the conversation
	chatMessages, err := database.GetAllMessages(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Convert ChatMessage to models.Message for AI request
	messages := make([]models.Message, 0, len(chatMessages))
	for _, msg := range chatMessages {
		messages = append(messages, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages, nil
}

// BuildContextWithSystemPrompt builds context including an optional system prompt
// Requirements: 2.3
func (s *ChatService) BuildContextWithSystemPrompt(conversationID int64, systemPrompt string) ([]models.Message, error) {
	messages, err := s.BuildContext(conversationID)
	if err != nil {
		return nil, err
	}

	// Prepend system prompt if provided
	if systemPrompt != "" {
		systemMsg := models.Message{
			Role:    "system",
			Content: systemPrompt,
		}
		messages = append([]models.Message{systemMsg}, messages...)
	}

	return messages, nil
}

// SendMessage sends a user message and streams the AI response
// Requirements: 2.1-2.6 - Route to appropriate provider based on model
// Requirements: 2.3 - Include all previous messages as context
// Requirements: 6.2 - Check user balance before AI call
// Requirements: 10.1-10.5 - Handle provider-specific errors
func (s *ChatService) SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	// Validate request
	if req.Content == "" {
		return nil, ErrEmptyMessage
	}

	// Check user balance before proceeding (Requirements: 6.2)
	balance, err := database.GetUserBalance(req.UserID)
	if err != nil {
		if errors.Is(err, database.ErrBalanceNotFound) {
			// Auto-create balance for users who don't have one
			balance, err = database.CreateUserBalance(req.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to create user balance: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get user balance: %w", err)
		}
	}

	// Check if user has sufficient balance (minimum $0.001 required)
	const minRequiredBalance = 0.001
	if balance.Balance < minRequiredBalance {
		return nil, ErrInsufficientBalance
	}

	// Check if balance status is exhausted
	if balance.Status == database.BalanceStatusExhausted {
		return nil, ErrInsufficientBalance
	}

	// Verify conversation exists and belongs to user
	conv, err := database.GetConversation(req.ConversationID, req.UserID)
	if err != nil {
		if errors.Is(err, database.ErrConversationNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Determine which model to use
	model := conv.Model
	if req.Model != "" {
		model = req.Model
	}

	// Generate request ID for logging
	requestID := fmt.Sprintf("chat-%d-%d", req.ConversationID, req.UserID)

	// Log the model being used for debugging
	logrus.WithFields(logrus.Fields{
		"conversation_id":    req.ConversationID,
		"conversation_model": conv.Model,
		"request_model":      req.Model,
		"final_model":        model,
		"request_id":         requestID,
	}).Info("Chat request model selection")

	// Save user message to database first (Requirements: 2.1)
	userMessage, err := database.CreateMessage(req.ConversationID, "user", req.Content, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Build context with all previous messages (Requirements: 2.3)
	contextMessages, err := s.BuildContextWithSystemPrompt(req.ConversationID, conv.SystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Try to use ProviderRouter if available (Requirements: 2.1-2.6)
	if s.providerRouter != nil {
		return s.sendMessageWithProvider(ctx, model, contextMessages, userMessage, requestID)
	}

	// Fallback to legacy CursorService if ProviderRouter not configured
	return s.sendMessageWithCursor(ctx, model, contextMessages, userMessage)
}

// sendMessageWithProvider sends message using the ProviderRouter
// Requirements: 2.1-2.6, 10.1-10.5
func (s *ChatService) sendMessageWithProvider(ctx context.Context, model string, messages []models.Message, userMessage *models.ChatMessage, requestID string) (*SendMessageResponse, error) {
	// Get the appropriate provider for the model (Requirements: 2.1-2.5)
	provider, err := s.providerRouter.GetProvider(model)
	if err != nil {
		// Requirements: 2.6 - Return PROVIDER_NOT_AVAILABLE error
		return nil, mapProviderError(err, "unknown", model, requestID)
	}

	providerName := provider.GetProviderName()

	logrus.WithFields(logrus.Fields{
		"model":      model,
		"provider":   providerName,
		"request_id": requestID,
	}).Info("Routing request to provider")

	// Create chat request for provider
	chatRequest := &models.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	// Send to provider
	streamChan, err := provider.ChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, mapProviderError(err, providerName, model, requestID)
	}

	return &SendMessageResponse{
		UserMessage: userMessage,
		StreamChan:  streamChan,
	}, nil
}

// sendMessageWithCursor sends message using the legacy CursorService
func (s *ChatService) sendMessageWithCursor(ctx context.Context, model string, messages []models.Message, userMessage *models.ChatMessage) (*SendMessageResponse, error) {
	// Create chat completion request
	chatRequest := &models.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	// Send to AI service
	cursorStreamChan, _, err := s.cursorService.ChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send to AI: %w", err)
	}

	// Convert cursor stream to unified StreamEvent format
	eventChan := make(chan models.StreamEvent)
	go convertCursorStreamToEvents(cursorStreamChan, eventChan)

	return &SendMessageResponse{
		UserMessage: userMessage,
		StreamChan:  eventChan,
	}, nil
}

// convertCursorStreamToEvents converts legacy cursor stream to unified StreamEvent format
func convertCursorStreamToEvents(cursorChan <-chan interface{}, eventChan chan<- models.StreamEvent) {
	defer close(eventChan)

	// Use the CursorProvider's conversion logic
	cursorProvider := providers.NewCursorProvider(nil)
	_ = cursorProvider // We'll implement inline conversion here

	hasStarted := false
	var totalUsage *models.TokenUsage

	for event := range cursorChan {
		if !hasStarted {
			eventChan <- models.StreamEvent{Type: "start"}
			hasStarted = true
		}

		switch v := event.(type) {
		case string:
			// Try to parse as JSON
			var cursorEvent models.CursorEventData
			if err := parseJSON(v, &cursorEvent); err == nil {
				switch cursorEvent.Type {
				case "delta":
					if cursorEvent.Delta != "" {
						eventChan <- models.StreamEvent{Type: "content", Content: cursorEvent.Delta}
					}
				case "error":
					eventChan <- models.StreamEvent{Type: "error", Error: cursorEvent.ErrorText}
					return
				case "done":
					if cursorEvent.MessageMetadata != nil && cursorEvent.MessageMetadata.Usage != nil {
						totalUsage = &models.TokenUsage{
							PromptTokens:     cursorEvent.MessageMetadata.Usage.InputTokens,
							CompletionTokens: cursorEvent.MessageMetadata.Usage.OutputTokens,
							TotalTokens:      cursorEvent.MessageMetadata.Usage.TotalTokens,
						}
					}
				}
			} else {
				// Plain text content
				eventChan <- models.StreamEvent{Type: "content", Content: v}
			}
		case models.CursorEventData:
			switch v.Type {
			case "delta":
				if v.Delta != "" {
					eventChan <- models.StreamEvent{Type: "content", Content: v.Delta}
				}
			case "error":
				eventChan <- models.StreamEvent{Type: "error", Error: v.ErrorText}
				return
			case "done":
				if v.MessageMetadata != nil && v.MessageMetadata.Usage != nil {
					totalUsage = &models.TokenUsage{
						PromptTokens:     v.MessageMetadata.Usage.InputTokens,
						CompletionTokens: v.MessageMetadata.Usage.OutputTokens,
						TotalTokens:      v.MessageMetadata.Usage.TotalTokens,
					}
				}
			}
		case error:
			eventChan <- models.StreamEvent{Type: "error", Error: v.Error()}
			return
		}
	}

	if totalUsage != nil {
		eventChan <- models.StreamEvent{Type: "usage", Tokens: totalUsage}
	}
	eventChan <- models.StreamEvent{Type: "done"}
}

// parseJSON is a helper to parse JSON strings
func parseJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// GetConversationMessages retrieves messages for a conversation with pagination
// Requirements: 1.3, 7.2
func (s *ChatService) GetConversationMessages(conversationID, userID int64, page, limit int) ([]models.ChatMessage, int, error) {
	// Verify conversation belongs to user
	belongs, err := database.ConversationBelongsToUser(conversationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to verify conversation ownership: %w", err)
	}
	if !belongs {
		return nil, 0, ErrUnauthorized
	}

	return database.GetMessages(conversationID, page, limit)
}

// SaveAssistantMessage saves the AI response to the database
// Requirements: 2.4 - Save response with token usage information
func (s *ChatService) SaveAssistantMessage(conversationID int64, content string, tokens int, cost float64) (*models.ChatMessage, error) {
	return database.CreateMessage(conversationID, "assistant", content, tokens, cost)
}

// GetAvailableModels returns the list of available AI models
// Requirements: 3.1
func (s *ChatService) GetAvailableModels() []string {
	return s.config.GetModels()
}
