package services

import (
	"context"
	"Curry2API-go/config"
	"Curry2API-go/middleware"
	"Curry2API-go/models"
	"Curry2API-go/utils"
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const cursorAPIURL = "https://cursor.com/api/chat"

// CursorService 主服务协调器
// 协调认证、HTTP 客户端和消息处理模块
type CursorService struct {
	config *config.Config
	client *req.Client
	mainJS string
	envJS  string

	// 子模块
	auth    *authManager
	http    *httpClient
	message *messageHandler
}

// NewCursorService 创建新的 CursorService 实例
func NewCursorService(cfg *config.Config) *CursorService {
	// 读取 JavaScript 文件
	mainJS, err := os.ReadFile(filepath.Join("jscode", "main.js"))
	if err != nil {
		logrus.Fatalf("failed to read jscode/main.js: %v", err)
	}

	envJS, err := os.ReadFile(filepath.Join("jscode", "env.js"))
	if err != nil {
		logrus.Fatalf("failed to read jscode/env.js: %v", err)
	}

	// 创建 HTTP 客户端
	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.Warnf("failed to create cookie jar: %v", err)
	}

	client := req.C()
	client.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	client.ImpersonateChrome()
	if jar != nil {
		client.SetCookieJar(jar)
	}

	service := &CursorService{
		config: cfg,
		client: client,
		mainJS: string(mainJS),
		envJS:  string(envJS),
	}

	// 初始化子模块
	service.auth = &authManager{service: service}
	service.http = &httpClient{service: service}
	service.message = &messageHandler{service: service}

	return service
}

// mapToCursorModel 将配置中的模型名称映射到Cursor API支持的格式
func mapToCursorModel(model string) string {
	// Cursor API 模型名称映射
	cursorModelMap := map[string]string{
		// GPT-5.2 系列
		"gpt-5.2":            "gpt-5.2",
		
		// Claude 系列
		"claude-3.5-sonnet":  "claude-3.5-sonnet",
		"claude-3.5-haiku":   "claude-3.5-haiku",
		"claude-3.7-sonnet":  "claude-3.7-sonnet",
		"claude-4-sonnet":    "claude-4-sonnet",
		"claude-4.5-sonnet":  "claude-4.5-sonnet",
		"claude-4-opus":      "claude-4-opus",
		"claude-4.1-opus":    "claude-4.1-opus",
		"claude-4.5-opus":    "claude-4.5-opus",
		"claude-4.5-haiku":   "claude-4.5-haiku",
		
		// GPT 系列
		"gpt-5":              "gpt-5",
		"gpt-5.1":            "gpt-5.1",
		"gpt-5-codex":        "gpt-5-codex",
		"gpt-5.1-codex":      "gpt-5.1-codex",
		"gpt-5.1-codex-max":  "gpt-5.1-codex-max",
		"gpt-5-mini":         "gpt-5-mini",
		"gpt-5-nano":         "gpt-5-nano",
		"gpt-4.1":            "gpt-4.1",
		"gpt-4o":             "gpt-4o",
		
		// O 系列
		"o3":                 "o3",
		"o4-mini":            "o4-mini",
		
		// Gemini 系列
		"gemini-2.5-pro":     "gemini-2.5-pro",
		"gemini-2.5-flash":   "gemini-2.5-flash",
		"gemini-3-pro-preview": "gemini-3-pro-preview",
		
		// 其他模型
		"deepseek-r1":        "deepseek-r1",
		"deepseek-v3.1":      "deepseek-v3.1",
		"kimi-k2-instruct":   "kimi-k2-instruct",
		"grok-3":             "grok-3",
		"grok-3-mini":        "grok-3-mini",
		"grok-4":             "grok-4",
		"code-supernova-1-million": "code-supernova-1-million",
	}
	
	if cursorModel, exists := cursorModelMap[model]; exists {
		return cursorModel
	}
	
	// 如果没有映射，返回原始名称
	return model
}

// ChatCompletion 创建聊天完成流
// 这是主要的对外接口，协调各个子模块完成请求
func (s *CursorService) ChatCompletion(ctx context.Context, request *models.ChatCompletionRequest) (<-chan interface{}, *middleware.CursorSessionInfo, error) {
	// 1. 消息处理：截断和转换
	truncatedMessages := s.message.truncateMessages(request.Messages)
	cursorMessages := models.ToCursorMessages(truncatedMessages, s.config.SystemPromptInject)

	// 映射模型名称到Cursor API格式
	cursorModel := mapToCursorModel(request.Model)
	logrus.WithFields(logrus.Fields{
		"request_model": request.Model,
		"cursor_model":  cursorModel,
	}).Info("ChatCompletion model mapping")

	// 2. 构建请求 payload
	payload := models.CursorRequest{
		Context:  []interface{}{},
		Model:    cursorModel,
		ID:       utils.GenerateRandomString(16),
		Messages: cursorMessages,
		Trigger:  "submit-message",
		Tools:    request.Tools, // 传递工具定义
	}
	
	// 记录工具信息
	if len(request.Tools) > 0 {
		logrus.WithField("tool_count", len(request.Tools)).Debug("Passing tools to Cursor API")
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal cursor payload: %w", err)
	}
	
	// Log the payload for debugging
	logrus.WithFields(logrus.Fields{
		"model":         payload.Model,
		"message_count": len(payload.Messages),
		"payload_size":  len(jsonPayload),
	}).Info("Sending request to Cursor API")

	// 3. 认证：获取 x-is-human 令牌
	xIsHuman, err := s.auth.fetchXIsHuman(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 4. HTTP 请求：发送到 Cursor API（返回使用的 session）
	resp, session, err := s.http.sendChatRequest(ctx, xIsHuman, jsonPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("cursor request failed: %w", err)
	}

	// 5. 流处理：启动 SSE 消费协程
	output := make(chan interface{}, 32)
	go s.http.consumeSSE(ctx, resp, output, session)

	return output, session, nil
}
