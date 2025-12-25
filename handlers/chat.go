package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"Curry2API-go/config"
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ChatHandler handles all chat-related HTTP requests
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.4, 3.1, 7.2, 7.3, 11.1-11.5
type ChatHandler struct {
	chatService    *services.ChatService
	providerRouter *services.ProviderRouter
	config         *config.Config
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(chatService *services.ChatService, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		config:      cfg,
	}
}

// NewChatHandlerWithRouter creates a new ChatHandler instance with ProviderRouter
// Requirements: 11.1-11.5
func NewChatHandlerWithRouter(chatService *services.ChatService, providerRouter *services.ProviderRouter, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		providerRouter: providerRouter,
		config:         cfg,
	}
}

// SetProviderRouter sets the provider router for the handler
func (h *ChatHandler) SetProviderRouter(router *services.ProviderRouter) {
	h.providerRouter = router
}

// Request/Response types for chat handlers

// CreateConversationRequest represents the request body for creating a conversation
type CreateConversationRequest struct {
	Title        string `json:"title"`
	Model        string `json:"model" binding:"required"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// UpdateConversationRequest represents the request body for updating a conversation
type UpdateConversationRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
	Model   string `json:"model,omitempty"` // Optional: override conversation model
}

// CreateConversation creates a new chat conversation
// POST /api/chat/conversations
// Requirements: 1.1
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate model
	if !h.config.IsValidModel(req.Model) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid model specified: "+req.Model,
			"validation_error",
			"invalid_model",
		))
		return
	}

	// Set default title if not provided
	title := req.Title
	if title == "" {
		title = "新对话"
	}

	// Create conversation in database
	conv, err := database.CreateConversation(userID, title, req.Model)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to create conversation")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to create conversation",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conv,
	})
}

// GetConversations retrieves paginated conversations for the current user
// GET /api/chat/conversations
// Query params: page (default 1), limit (default 20, max 100)
// Requirements: 1.2, 7.3
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	// Get conversations from database
	conversations, total, err := database.GetConversations(userID, page, limit)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get conversations")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve conversations",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversations": conversations,
			"total":         total,
			"page":          page,
			"limit":         limit,
		},
	})
}

// GetConversation retrieves a single conversation by ID
// GET /api/chat/conversations/:id
// Requirements: 1.3
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse conversation ID
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid conversation ID",
			"validation_error",
			"invalid_id",
		))
		return
	}

	// Get conversation from database
	conv, err := database.GetConversation(convID, userID)
	if err != nil {
		if err == database.ErrConversationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Conversation not found",
				"not_found",
				"conversation_not_found",
			))
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to get conversation")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve conversation",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conv,
	})
}

// UpdateConversation updates a conversation's title and/or model
// PUT /api/chat/conversations/:id
// Requirements: 1.5
func (h *ChatHandler) UpdateConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse conversation ID
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid conversation ID",
			"validation_error",
			"invalid_id",
		))
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Get existing conversation to preserve unchanged fields
	existingConv, err := database.GetConversation(convID, userID)
	if err != nil {
		if err == database.ErrConversationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Conversation not found",
				"not_found",
				"conversation_not_found",
			))
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to get conversation for update")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve conversation",
			"internal_error",
			"database_error",
		))
		return
	}

	// Use existing values if not provided in request
	title := req.Title
	if title == "" {
		title = existingConv.Title
	}

	model := req.Model
	if model == "" {
		model = existingConv.Model
	} else {
		// Validate model if provided
		if !h.config.IsValidModel(model) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid model specified: "+model,
				"validation_error",
				"invalid_model",
			))
			return
		}
	}

	// Update conversation in database
	err = database.UpdateConversation(convID, userID, title, model)
	if err != nil {
		if err == database.ErrConversationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Conversation not found",
				"not_found",
				"conversation_not_found",
			))
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to update conversation")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to update conversation",
			"internal_error",
			"database_error",
		))
		return
	}

	// Get updated conversation to return
	updatedConv, err := database.GetConversation(convID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to get updated conversation")
		// Still return success since update was successful
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Conversation updated successfully",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedConv,
	})
}

// DeleteConversation deletes a conversation and all its messages
// DELETE /api/chat/conversations/:id
// Requirements: 1.4
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse conversation ID
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid conversation ID",
			"validation_error",
			"invalid_id",
		))
		return
	}

	// Delete conversation from database (cascade deletes messages)
	err = database.DeleteConversation(convID, userID)
	if err != nil {
		if err == database.ErrConversationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Conversation not found",
				"not_found",
				"conversation_not_found",
			))
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to delete conversation")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to delete conversation",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Conversation deleted successfully",
	})
}

// GetMessages retrieves paginated messages for a conversation
// GET /api/chat/conversations/:id/messages
// Query params: page (default 1), limit (default 50, max 100)
// Requirements: 1.3, 7.2
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse conversation ID
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid conversation ID",
			"validation_error",
			"invalid_id",
		))
		return
	}

	// Verify conversation belongs to user
	belongs, err := database.ConversationBelongsToUser(convID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to verify conversation ownership")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to verify conversation ownership",
			"internal_error",
			"database_error",
		))
		return
	}
	if !belongs {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"Conversation not found",
			"not_found",
			"conversation_not_found",
		))
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	// Get messages from database
	messages, total, err := database.GetMessages(convID, page, limit)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Error("Failed to get messages")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve messages",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"messages": messages,
			"total":    total,
			"page":     page,
			"limit":    limit,
		},
	})
}


// SendMessage sends a message and streams the AI response via SSE
// POST /api/chat/conversations/:id/messages
// Requirements: 2.1, 2.2, 2.4, 2.5
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse conversation ID
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id":  userID,
			"param_id": c.Param("id"),
		}).Warn("Invalid conversation ID format")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid conversation ID",
			"validation_error",
			"invalid_id",
		))
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Warn("Invalid request format for send message")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate content is not empty
	if strings.TrimSpace(req.Content) == "" {
		logrus.WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
		}).Warn("Empty message content")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Message content cannot be empty",
			"validation_error",
			"empty_content",
		))
		return
	}

	// Validate model if provided
	if req.Model != "" && !h.config.IsValidModel(req.Model) {
		logrus.WithFields(logrus.Fields{
			"user_id":         userID,
			"conversation_id": convID,
			"model":           req.Model,
		}).Warn("Invalid model specified")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid model specified: "+req.Model,
			"validation_error",
			"invalid_model",
		))
		return
	}

	// Send message using chat service
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	response, err := h.chatService.SendMessage(ctx, services.SendMessageRequest{
		ConversationID: convID,
		UserID:         userID,
		Content:        req.Content,
		Model:          req.Model,
	})
	if err != nil {
		h.handleSendMessageError(c, err, userID, convID)
		return
	}

	// Set up SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Send start event with user message ID
	startEvent := models.ChatStreamEvent{
		Type:      "start",
		MessageID: response.UserMessage.ID,
	}
	sendSSEEvent(c, startEvent)

	// Stream AI response
	var fullContent strings.Builder
	var totalPromptTokens, totalCompletionTokens int

	for event := range response.StreamChan {
		select {
		case <-ctx.Done():
			// Context cancelled or timeout, send error event
			// Requirements: 2.5 - Handle stream errors gracefully
			var errorMsg string
			if ctx.Err() == context.DeadlineExceeded {
				errorMsg = "Request timed out. Please try again."
				logrus.WithFields(logrus.Fields{
					"user_id":         userID,
					"conversation_id": convID,
				}).Warn("Chat stream timeout")
			} else {
				errorMsg = "Request was cancelled"
				logrus.WithFields(logrus.Fields{
					"user_id":         userID,
					"conversation_id": convID,
				}).Info("Chat stream cancelled by client")
			}
			errorEvent := models.ChatStreamEvent{
				Type:  "error",
				Error: errorMsg,
			}
			sendSSEEvent(c, errorEvent)
			return
		default:
			// Process unified StreamEvent format
			// Requirements: 2.5 - Handle stream errors gracefully
			// Requirements: 9.1, 9.4, 9.5 - Token usage and cost tracking
			switch event.Type {
			case "start":
				// Start event - already sent start event above
				continue
			case "content":
				// Content delta
				fullContent.WriteString(event.Content)
				contentEvent := models.ChatStreamEvent{
					Type:  "content",
					Delta: event.Content,
				}
				sendSSEEvent(c, contentEvent)
			case "usage":
				// Token usage information (Requirements: 9.1)
				if event.Tokens != nil {
					totalPromptTokens = event.Tokens.PromptTokens
					totalCompletionTokens = event.Tokens.CompletionTokens
				}
			case "error":
				// Error event
				logrus.WithFields(logrus.Fields{
					"user_id":         userID,
					"conversation_id": convID,
					"error":           event.Error,
				}).Error("AI service returned error during streaming")
				errorEvent := models.ChatStreamEvent{
					Type:  "error",
					Error: event.Error,
				}
				sendSSEEvent(c, errorEvent)
				return
			case "done":
				// Done event - will be handled after loop
				continue
			}
		}
	}

	// Get conversation to retrieve model info for billing
	conv, convErr := database.GetConversation(convID, userID)
	model := ""
	if convErr == nil {
		model = conv.Model
	}
	// Use request model if provided
	if req.Model != "" {
		model = req.Model
	}

	// Save assistant message to database (Requirements: 2.4)
	totalTokens := totalPromptTokens + totalCompletionTokens
	// Calculate cost using pricing service (Requirements: 9.3)
	cost := services.CalculateCost(model, totalPromptTokens, totalCompletionTokens)
	if cost == 0 {
		// Fallback to default calculation if model not in pricing table
		cost = calculateCost(totalPromptTokens, totalCompletionTokens)
	}

	assistantMsg, err := h.chatService.SaveAssistantMessage(convID, fullContent.String(), totalTokens, cost)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"conversation_id": convID,
		}).Error("Failed to save assistant message")
		// Still send done event even if save fails
	}

	// Deduct balance after AI response (Requirements: 6.1)
	if totalTokens > 0 {
		_, deductErr := database.DeductBalance(userID, totalTokens, "chat", model)
		if deductErr != nil {
			logrus.WithError(deductErr).WithFields(logrus.Fields{
				"user_id":         userID,
				"conversation_id": convID,
				"tokens":          totalTokens,
				"cost":            cost,
			}).Error("Failed to deduct balance for chat usage")
			// Don't fail the request, just log the error
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id":         userID,
				"conversation_id": convID,
				"tokens":          totalTokens,
				"cost":            cost,
				"model":           model,
			}).Info("Balance deducted for chat usage")
		}
	}

	// Create usage record for chat interaction (Requirements: 6.3, 9.5)
	if totalTokens > 0 {
		// Get username for usage record
		username := ""
		user, userErr := database.GetUserByID(userID)
		if userErr == nil && user != nil {
			username = user.Username
		}

		// Determine provider from model name for usage record (Requirements: 9.5)
		provider := services.GetProviderFromModel(model)

		now := time.Now()
		usageRecord := &database.UsageRecord{
			UserID:           userID,
			Username:         username,
			APIToken:         "chat",
			TokenName:        fmt.Sprintf("Online Chat (%s)", provider),
			Model:            model,
			PromptTokens:     totalPromptTokens,
			CompletionTokens: totalCompletionTokens,
			TotalTokens:      totalTokens,
			CursorSession:    "",
			StatusCode:       200,
			ErrorMessage:     "",
			RequestTime:      now,
			ResponseTime:     now,
			DurationMs:       0,
		}

		if insertErr := database.InsertUsageRecord(usageRecord); insertErr != nil {
			logrus.WithError(insertErr).WithFields(logrus.Fields{
				"user_id":         userID,
				"conversation_id": convID,
				"tokens":          totalTokens,
			}).Error("Failed to create usage record for chat")
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id":         userID,
				"conversation_id": convID,
				"tokens":          totalTokens,
				"model":           model,
				"provider":        provider,
				"cost":            cost,
			}).Debug("Usage record created for chat")
		}
	}

	// Send done event with token usage
	doneEvent := models.ChatStreamEvent{
		Type: "done",
		Tokens: &models.ChatTokenUsage{
			Prompt:     totalPromptTokens,
			Completion: totalCompletionTokens,
		},
		Cost: cost,
	}
	if assistantMsg != nil {
		doneEvent.MessageID = assistantMsg.ID
	}
	sendSSEEvent(c, doneEvent)
}

// ModelResponse represents a model in the API response
type ModelResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Provider      string  `json:"provider"`
	ContextWindow int     `json:"context_window"`
	InputPrice    float64 `json:"input_price"`
	OutputPrice   float64 `json:"output_price"`
	IsAvailable   bool    `json:"is_available"`
}

// ProviderModelsResponse represents models grouped by provider
type ProviderModelsResponse struct {
	Provider string          `json:"provider"`
	Models   []ModelResponse `json:"models"`
}

// GetModels returns the list of available AI models
// GET /api/chat/models
// Requirements: 11.1, 11.2, 11.3, 11.4, 11.5
func (h *ChatHandler) GetModels(c *gin.Context) {
	// If ProviderRouter is available, use it to get models from all providers
	if h.providerRouter != nil {
		h.getModelsFromProviderRouter(c)
		return
	}

	// Fallback to legacy behavior using config models
	h.getModelsFromConfig(c)
}

// getModelsFromProviderRouter returns models from all configured providers
// Requirements: 11.1, 11.2, 11.3, 11.4, 11.5
func (h *ChatHandler) getModelsFromProviderRouter(c *gin.Context) {
	// Get all models from provider router
	allModels := h.providerRouter.GetAllModels()

	// Group models by provider (Requirements: 11.4)
	providerModels := make(map[string][]ModelResponse)
	providerOrder := []string{} // Track order of providers

	for _, model := range allModels {
		modelResp := ModelResponse{
			ID:            model.ID,
			Name:          model.Name,
			Provider:      model.Provider,
			ContextWindow: model.ContextWindow,
			InputPrice:    model.InputPrice,
			OutputPrice:   model.OutputPrice,
			IsAvailable:   model.IsAvailable, // Requirements: 11.3, 11.5
		}

		if _, exists := providerModels[model.Provider]; !exists {
			providerOrder = append(providerOrder, model.Provider)
		}
		providerModels[model.Provider] = append(providerModels[model.Provider], modelResp)
	}

	// Build grouped response
	groupedModels := make([]ProviderModelsResponse, 0, len(providerModels))
	for _, provider := range providerOrder {
		groupedModels = append(groupedModels, ProviderModelsResponse{
			Provider: provider,
			Models:   providerModels[provider],
		})
	}

	// Also return flat list for backward compatibility
	flatModels := make([]ModelResponse, 0, len(allModels))
	for _, model := range allModels {
		flatModels = append(flatModels, ModelResponse{
			ID:            model.ID,
			Name:          model.Name,
			Provider:      model.Provider,
			ContextWindow: model.ContextWindow,
			InputPrice:    model.InputPrice,
			OutputPrice:   model.OutputPrice,
			IsAvailable:   model.IsAvailable,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"models":          flatModels,      // Flat list for backward compatibility
			"models_grouped":  groupedModels,   // Grouped by provider (Requirements: 11.4)
		},
	})
}

// getModelsFromConfig returns models from config (legacy fallback)
func (h *ChatHandler) getModelsFromConfig(c *gin.Context) {
	modelNames := h.config.GetModels()
	modelList := make([]gin.H, 0, len(modelNames))

	for _, modelID := range modelNames {
		// Get model configuration info
		modelConfig, exists := models.GetModelConfig(modelID)

		modelInfo := gin.H{
			"id":           modelID,
			"name":         modelID, // Use model ID as display name
			"provider":     "Unknown",
			"is_available": true, // Legacy models are always available
		}

		// Add config info if available
		if exists {
			modelInfo["provider"] = modelConfig.Provider
			modelInfo["max_tokens"] = modelConfig.MaxTokens
			modelInfo["context_window"] = modelConfig.ContextWindow
		}

		// Get pricing info
		pricing := services.GetModelPricing(modelID)
		if pricing != nil {
			modelInfo["input_price"] = pricing.InputPrice
			modelInfo["output_price"] = pricing.OutputPrice
		}

		modelList = append(modelList, modelInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"models": modelList,
		},
	})
}

// Helper functions

// handleSendMessageError handles errors from SendMessage and returns appropriate HTTP responses
// Requirements: 2.5, 10.1-10.5 - Display error message and allow retry
func (h *ChatHandler) handleSendMessageError(c *gin.Context, err error, userID, convID int64) {
	logFields := logrus.Fields{
		"user_id":         userID,
		"conversation_id": convID,
		"error":           err.Error(),
	}

	switch {
	case err == services.ErrConversationNotFound:
		logrus.WithFields(logFields).Warn("Conversation not found")
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"Conversation not found",
			"not_found",
			"conversation_not_found",
		))

	case err == services.ErrEmptyMessage:
		logrus.WithFields(logFields).Warn("Empty message content")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Message content cannot be empty",
			"validation_error",
			"empty_content",
		))

	case err == services.ErrInsufficientBalance:
		// Requirements: 6.2 - Return 402 error if insufficient balance
		logrus.WithFields(logFields).Info("Insufficient balance for chat")
		c.JSON(http.StatusPaymentRequired, models.NewErrorResponse(
			"Insufficient balance. Please recharge your account to continue.",
			"payment_required",
			"insufficient_balance",
		))

	case err == services.ErrAIServiceUnavailable:
		logrus.WithFields(logFields).Error("AI service unavailable")
		c.JSON(http.StatusBadGateway, models.NewErrorResponse(
			"AI service is temporarily unavailable. Please try again later.",
			"service_unavailable",
			"ai_service_unavailable",
		))

	case err == services.ErrAIServiceTimeout:
		logrus.WithFields(logFields).Error("AI service timeout")
		c.JSON(http.StatusGatewayTimeout, models.NewErrorResponse(
			"AI service request timed out. Please try again.",
			"timeout",
			"ai_service_timeout",
		))

	case err == services.ErrInvalidModel:
		logrus.WithFields(logFields).Warn("Invalid model specified")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid model specified",
			"validation_error",
			"invalid_model",
		))

	case err == services.ErrUnauthorized:
		logrus.WithFields(logFields).Warn("Unauthorized access to conversation")
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"You do not have access to this conversation",
			"forbidden",
			"unauthorized_access",
		))

	// Provider-specific errors (Requirements: 10.1-10.5)
	case err == services.ErrProviderNotAvailable:
		logrus.WithFields(logFields).Warn("Provider not available")
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse(
			"The selected AI provider is not available. Please configure the API key or choose a different model.",
			"provider_not_available",
			"PROVIDER_NOT_AVAILABLE",
		))

	case err == services.ErrInvalidAPIKey:
		// Requirements: 10.1 - Handle 401 errors
		logrus.WithFields(logFields).Error("Invalid API key")
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"API key is invalid or expired. Please contact administrator.",
			"invalid_api_key",
			"INVALID_API_KEY",
		))

	case err == services.ErrRateLimited:
		// Requirements: 10.2 - Handle 429 errors
		logrus.WithFields(logFields).Warn("Rate limited by provider")
		c.JSON(http.StatusTooManyRequests, models.NewErrorResponse(
			"Rate limit exceeded, please try again later.",
			"rate_limited",
			"RATE_LIMITED",
		))

	case err == services.ErrProviderError:
		// Requirements: 10.3 - Handle 500-599 errors
		logrus.WithFields(logFields).Error("Provider error")
		c.JSON(http.StatusBadGateway, models.NewErrorResponse(
			"AI service temporarily unavailable. Please try again later.",
			"provider_error",
			"PROVIDER_ERROR",
		))

	case err == services.ErrTimeout:
		// Requirements: 10.4 - Handle timeout errors
		logrus.WithFields(logFields).Error("Provider timeout")
		c.JSON(http.StatusGatewayTimeout, models.NewErrorResponse(
			"Request timed out. Please try again.",
			"timeout",
			"TIMEOUT",
		))

	case err == services.ErrContextTooLong:
		// Requirements: 10.5 - Handle context length errors
		logrus.WithFields(logFields).Warn("Context too long")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Message too long for this model. Please reduce the conversation length.",
			"context_too_long",
			"CONTEXT_TOO_LONG",
		))

	default:
		// Generic error - log full details for debugging
		logrus.WithError(err).WithFields(logFields).Error("Failed to send message")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to send message. Please try again.",
			"internal_error",
			"ai_service_error",
		))
	}
}

// sendSSEEvent sends a Server-Sent Event to the client
func sendSSEEvent(c *gin.Context, event models.ChatStreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal SSE event")
		return
	}

	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.(http.Flusher).Flush()
}

// calculateCost calculates the cost based on token usage
// This is a simplified calculation - in production, use model-specific pricing
func calculateCost(promptTokens, completionTokens int) float64 {
	// Default pricing: $0.01 per 1K prompt tokens, $0.03 per 1K completion tokens
	promptCost := float64(promptTokens) / 1000.0 * 0.01
	completionCost := float64(completionTokens) / 1000.0 * 0.03
	return promptCost + completionCost
}

// streamResponseFromChannel reads from the AI response channel and streams to client
func streamResponseFromChannel(c *gin.Context, streamChan <-chan interface{}, convID int64, chatService *services.ChatService) {
	var fullContent strings.Builder
	var totalPromptTokens, totalCompletionTokens int

	// Create a buffered writer for SSE
	writer := bufio.NewWriter(c.Writer)
	defer writer.Flush()

	for chunk := range streamChan {
		switch v := chunk.(type) {
		case string:
			fullContent.WriteString(v)
			event := models.ChatStreamEvent{
				Type:  "content",
				Delta: v,
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(writer, "data: %s\n\n", data)
			writer.Flush()
			c.Writer.(http.Flusher).Flush()

		case map[string]interface{}:
			if errMsg, ok := v["error"].(string); ok {
				event := models.ChatStreamEvent{
					Type:  "error",
					Error: errMsg,
				}
				data, _ := json.Marshal(event)
				fmt.Fprintf(writer, "data: %s\n\n", data)
				writer.Flush()
				c.Writer.(http.Flusher).Flush()
				return
			}
			if usage, ok := v["usage"].(map[string]interface{}); ok {
				if prompt, ok := usage["prompt_tokens"].(int); ok {
					totalPromptTokens = prompt
				}
				if completion, ok := usage["completion_tokens"].(int); ok {
					totalCompletionTokens = completion
				}
			}
		}
	}

	// Save assistant message
	totalTokens := totalPromptTokens + totalCompletionTokens
	cost := calculateCost(totalPromptTokens, totalCompletionTokens)

	assistantMsg, err := chatService.SaveAssistantMessage(convID, fullContent.String(), totalTokens, cost)
	if err != nil {
		logrus.WithError(err).Error("Failed to save assistant message")
	}

	// Send done event
	doneEvent := models.ChatStreamEvent{
		Type: "done",
		Tokens: &models.ChatTokenUsage{
			Prompt:     totalPromptTokens,
			Completion: totalCompletionTokens,
		},
		Cost: cost,
	}
	if assistantMsg != nil {
		doneEvent.MessageID = assistantMsg.ID
	}
	data, _ := json.Marshal(doneEvent)
	fmt.Fprintf(writer, "data: %s\n\n", data)
	writer.Flush()
	c.Writer.(http.Flusher).Flush()
}
