package handlers

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"Curry2API-go/config"
	"Curry2API-go/middleware"
	"Curry2API-go/models"
	"Curry2API-go/services"
	"Curry2API-go/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ClaudeHandler Claude API处理器
type ClaudeHandler struct {
	config            *config.Config
	cursorService     *services.CursorService
	openRouterService *services.OpenRouterService
	toolExecutor      *services.ToolExecutor
}

// NewClaudeHandler 创建新的Claude处理器
func NewClaudeHandler(cfg *config.Config) *ClaudeHandler {
	cursorService := services.NewCursorService(cfg)
	openRouterService := services.NewOpenRouterService(cfg)
	toolExecutor := services.NewToolExecutor()

	return &ClaudeHandler{
		config:            cfg,
		cursorService:     cursorService,
		openRouterService: openRouterService,
		toolExecutor:      toolExecutor,
	}
}

// ClaudeMessages 处理Claude Messages API请求
// POST /v1/messages
func (h *ClaudeHandler) ClaudeMessages(c *gin.Context) {
	// 读取原始请求体用于调试
	bodyBytes, _ := c.GetRawData()
	
	// 只在 Debug 级别记录完整请求体，避免日志过大
	if logrus.GetLevel() >= logrus.DebugLevel {
		// 截断过长的请求体
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 2000 {
			bodyStr = bodyStr[:2000] + "...[truncated]"
		}
		logrus.WithFields(logrus.Fields{
			"path": c.Request.URL.Path,
			"body": bodyStr,
		}).Debug("Received Claude request")
	}
	
	// 重新设置请求体，因为 GetRawData() 会消耗它
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	var request models.ClaudeMessageRequest
	
	// 绑定并验证JSON请求
	if err := c.ShouldBindJSON(&request); err != nil {
		logrus.WithError(err).Error("Failed to bind Claude request")
		errorResp := models.NewClaudeInvalidRequestError("Invalid request format: " + err.Error())
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// 如果未提供max_tokens，设置默认值（在验证之前）
	if request.MaxTokens == 0 {
		request.MaxTokens = 4096 // 默认值
		logrus.Debug("MaxTokens not provided, using default value: 4096")
	}

	// 验证请求字段
	if err := request.Validate(); err != nil {
		logrus.WithError(err).Error("Claude request validation failed")
		errorResp := models.NewClaudeInvalidRequestError(err.Error())
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// 先标准化模型名称，再验证
	normalizedModel := h.config.NormalizeModelName(request.Model)
	if !h.config.IsValidModel(normalizedModel) {
		logrus.WithFields(logrus.Fields{
			"model":            request.Model,
			"normalized_model": normalizedModel,
		}).Warn("Invalid model specified")
		errorResp := models.NewClaudeInvalidRequestError("Invalid model specified: " + request.Model)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Check token model access restriction
	// Requirements: 14.3
	apiKey, _ := c.Get("api_key")
	if apiKey != nil {
		km := middleware.GetKeyManager()
		if err := km.CheckTokenModelAccess(apiKey.(string), request.Model); err != nil {
			if err == middleware.ErrModelNotAllowed {
				logrus.WithFields(logrus.Fields{
					"model":   request.Model,
					"api_key": middleware.MaskKey(apiKey.(string)),
				}).Warn("Model access denied for token")
				errorResp := models.NewClaudeInvalidRequestError("Model not allowed - this token does not have access to model: " + request.Model)
				c.JSON(http.StatusForbidden, errorResp)
				return
			}
		}
	}

	// 使用标准化后的模型名称
	originalModel := request.Model
	request.Model = normalizedModel
	
	// 如果模型名称被标准化，记录日志
	if originalModel != request.Model {
		logrus.WithFields(logrus.Fields{
			"original_model":   originalModel,
			"normalized_model": request.Model,
		}).Debug("Model name normalized")
	}

	// 验证并调整max_tokens参数
	validatedMaxTokens := models.ValidateMaxTokens(request.Model, &request.MaxTokens)
	if validatedMaxTokens != nil {
		request.MaxTokens = *validatedMaxTokens
	}

	// 检查是否包含工具调用
	hasToolUse := h.toolExecutor.HasToolUse(&request)
	if hasToolUse {
		logrus.WithFields(logrus.Fields{
			"model":      request.Model,
			"tool_count": len(request.Tools),
		}).Info("Request contains tool definitions, injecting tool prompt")
		
		// 注入工具提示到请求中
		h.toolExecutor.InjectToolPrompt(&request)
		
		// 调试：打印注入后的系统提示类型
		logrus.WithFields(logrus.Fields{
			"system_type": fmt.Sprintf("%T", request.System),
		}).Debug("System prompt after tool injection")
	}

	// 转换Claude请求为OpenAI格式
	openAIRequest := request.ToOpenAIRequest()
	
	// 调试：打印转换后的系统消息
	if hasToolUse && len(openAIRequest.Messages) > 0 {
		for i, msg := range openAIRequest.Messages {
			if msg.Role == "system" {
				content := msg.GetStringContent()
				logrus.WithFields(logrus.Fields{
					"index":          i,
					"content_length": len(content),
					"has_tool_call":  strings.Contains(content, "<tool_call>"),
				}).Debug("OpenAI system message after conversion")
				break
			}
		}
	}
	
	// Capture request start time for usage tracking
	requestStartTime := time.Now()
	
	// Extract user and token info for usage tracking
	usageInfo, err := utils.ExtractUsageFromContext(c)
	if err != nil {
		logrus.WithError(err).Warn("Failed to extract usage context info for Claude API")
		// Continue processing - usage tracking is optional
	}
	
	// Store usage info and request details in context for downstream handlers
	c.Set("request_start_time", requestStartTime)
	c.Set("request_model", request.Model)
	if usageInfo != nil {
		c.Set("usage_info", usageInfo)
	}
	
	// Set the tracking function in context
	c.Set("track_usage_func", utils.UsageTrackingFunc(trackUsageFromContext))
	
	logrus.WithFields(logrus.Fields{
		"model":        request.Model,
		"stream":       request.Stream,
		"max_tokens":   request.MaxTokens,
		"messages":     len(request.Messages),
		"has_tool_use": hasToolUse,
	}).Info("Processing Claude API request")
	
	// 存储工具标记到上下文，供流处理器使用
	if hasToolUse {
		c.Set("has_tool_use", true)
	}

	// 检查是否为 OpenRouter 免费模型
	if services.IsOpenRouterModel(request.Model) {
		logrus.WithField("model", request.Model).Info("Using OpenRouter service for free model")
		
		chatGenerator, err := h.openRouterService.ChatCompletion(c.Request.Context(), openAIRequest)
		if err != nil {
			logrus.WithError(err).Error("Failed to create OpenRouter chat completion")
			errorResp := models.NewClaudeAPIError(err.Error())
			c.JSON(http.StatusInternalServerError, errorResp)
			return
		}
		
		// 设置 OpenRouter 标识
		c.Set("cursor_session", "openrouter-free-model")
		
		// 根据是否流式返回不同响应
		if request.Stream {
			utils.SafeClaudeStreamWrapper(utils.StreamClaudeCompletion, c, chatGenerator)
		} else {
			utils.NonStreamClaudeCompletion(c, chatGenerator)
		}
		return
	}

	// 调用Cursor服务（原有逻辑）
	chatGenerator, session, err := h.cursorService.ChatCompletion(c.Request.Context(), openAIRequest)
	if err != nil {
		h.handleCursorError(c, err)
		return
	}

	// 设置 session 信息
	h.setSessionInfo(c, session)

	// 根据是否流式返回不同响应
	if request.Stream {
		utils.SafeClaudeStreamWrapper(utils.StreamClaudeCompletion, c, chatGenerator)
	} else {
		utils.NonStreamClaudeCompletion(c, chatGenerator)
	}
}

// handleCursorError 处理 Cursor 服务错误
func (h *ClaudeHandler) handleCursorError(c *gin.Context, err error) {
	logrus.WithError(err).Error("Failed to create Claude chat completion")
	
	var errorResp *models.ClaudeErrorResponse
	
	switch e := err.(type) {
	case *middleware.CursorWebError:
		if e.StatusCode == http.StatusUnauthorized {
			errorResp = models.NewClaudeAuthenticationError(e.Message)
			c.JSON(http.StatusUnauthorized, errorResp)
		} else if e.StatusCode == http.StatusTooManyRequests {
			errorResp = models.NewClaudeRateLimitError(e.Message)
			c.JSON(http.StatusTooManyRequests, errorResp)
		} else {
			errorResp = models.NewClaudeAPIError(e.Message)
			c.JSON(e.StatusCode, errorResp)
		}
	case *middleware.AuthenticationError:
		errorResp = models.NewClaudeAuthenticationError(e.Message)
		c.JSON(http.StatusUnauthorized, errorResp)
	case *middleware.RateLimitError:
		errorResp = models.NewClaudeRateLimitError(e.Message)
		c.JSON(http.StatusTooManyRequests, errorResp)
	default:
		errorResp = models.NewClaudeAPIError(err.Error())
		c.JSON(http.StatusInternalServerError, errorResp)
	}
}

// setSessionInfo 设置 session 信息到上下文
func (h *ClaudeHandler) setSessionInfo(c *gin.Context, session *middleware.CursorSessionInfo) {
	if session != nil && session.Email != "" {
		c.Set("cursor_session", session.Email)
		logrus.Debugf("Claude API using Cursor session: %s", session.Email)
	} else {
		c.Set("cursor_session", "x-is-human-fallback")
		logrus.Debug("Claude API using x-is-human fallback method")
	}
}

// CountTokens 处理 Claude count_tokens API 请求
// POST /v1/messages/count_tokens
// 这是一个简化实现，返回估算的 token 数量
func (h *ClaudeHandler) CountTokens(c *gin.Context) {
	var request models.ClaudeMessageRequest
	
	// 绑定 JSON 请求
	if err := c.ShouldBindJSON(&request); err != nil {
		logrus.WithError(err).Debug("Failed to bind count_tokens request")
		errorResp := models.NewClaudeInvalidRequestError("Invalid request format: " + err.Error())
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}
	
	// 估算 token 数量（简单实现：每 4 个字符约 1 个 token）
	totalChars := 0
	
	// 计算系统提示的字符数
	if request.System != nil {
		switch sys := request.System.(type) {
		case string:
			totalChars += len(sys)
		case []interface{}:
			for _, item := range sys {
				if block, ok := item.(map[string]interface{}); ok {
					if text, exists := block["text"].(string); exists {
						totalChars += len(text)
					}
				}
			}
		}
	}
	
	// 计算消息的字符数
	for _, msg := range request.Messages {
		switch content := msg.Content.(type) {
		case string:
			totalChars += len(content)
		case []interface{}:
			for _, item := range content {
				if block, ok := item.(map[string]interface{}); ok {
					if text, exists := block["text"].(string); exists {
						totalChars += len(text)
					}
				}
			}
		}
	}
	
	// 计算工具定义的字符数
	for _, tool := range request.Tools {
		totalChars += len(tool.Name) + len(tool.Description)
	}
	
	// 估算 token 数量（每 4 个字符约 1 个 token，中文每 2 个字符约 1 个 token）
	// 这里使用保守估计
	estimatedTokens := (totalChars + 3) / 4
	if estimatedTokens < 1 {
		estimatedTokens = 1
	}
	
	// 返回 token 计数响应
	response := map[string]interface{}{
		"input_tokens": estimatedTokens,
	}
	
	c.JSON(http.StatusOK, response)
}
