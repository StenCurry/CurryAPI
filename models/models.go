package models

import (
	"encoding/json"
	"time"
)

// ChatCompletionRequest OpenAI聊天完成请求
type ChatCompletionRequest struct {
	Model        string    `json:"model" binding:"required"`
	Messages     []Message `json:"messages"` // 可选，Codex CLI 不使用
	Instructions string    `json:"instructions,omitempty"` // Codex CLI 使用此字段
	Stream       bool      `json:"stream,omitempty"`
	Temperature  *float64  `json:"temperature,omitempty"`
	MaxTokens    *int      `json:"max_tokens,omitempty"`
	TopP         *float64  `json:"top_p,omitempty"`
	Stop         []string  `json:"stop,omitempty"`
	User         string    `json:"user,omitempty"`
	Tools        []Tool    `json:"tools,omitempty"`        // 工具定义
	ToolChoice   interface{} `json:"tool_choice,omitempty"` // 工具选择策略
}

// Tool OpenAI工具定义
type Tool struct {
	Type     string              `json:"type"` // "function"
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition 函数定义
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Strict      bool                   `json:"strict,omitempty"`
}

// Message 消息结构
type Message struct {
	Role         string        `json:"role" binding:"required"`
	Content      interface{}   `json:"content" binding:"required"`
	ToolCallID   *string       `json:"tool_call_id,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用结构
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function 函数调用结构
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ContentPart 消息内容部分（用于多模态内容）
type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

// ChatCompletionResponse OpenAI聊天完成响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ChatCompletionStreamResponse 流式响应
type ChatCompletionStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// Choice 选择结构
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// StreamChoice 流式选择结构
type StreamChoice struct {
	Index        int            `json:"index"`
	Delta        StreamDelta    `json:"delta"`
	FinishReason *string        `json:"finish_reason"`
}

// StreamDelta 流式增量数据
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Model 模型信息
type Model struct {
	ID            string `json:"id"`
	Object        string `json:"object"`
	Created       int64  `json:"created"`
	OwnedBy       string `json:"owned_by"`
	MaxTokens     int    `json:"max_tokens,omitempty"`
	ContextWindow int    `json:"context_window,omitempty"`
}

// ModelsResponse 模型列表响应
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// CursorMessage Cursor消息格式
type CursorMessage struct {
	Role  string        `json:"role"`
	Parts []CursorPart  `json:"parts"`
}

// CursorPart Cursor消息部分
type CursorPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CursorRequest Cursor请求格式
type CursorRequest struct {
	Context  []interface{}   `json:"context"`
	Model    string          `json:"model"`
	ID       string          `json:"id"`
	Messages []CursorMessage `json:"messages"`
	Trigger  string          `json:"trigger"`
	Tools    []Tool          `json:"tools,omitempty"`    // 工具定义
}

// CursorEventData Cursor事件数据
type CursorEventData struct {
	Type            string                 `json:"type"`
	Delta           string                 `json:"delta,omitempty"`
	ErrorText       string                 `json:"errorText,omitempty"`
	MessageMetadata *CursorMessageMetadata `json:"messageMetadata,omitempty"`
}

// CursorMessageMetadata Cursor消息元数据
type CursorMessageMetadata struct {
	Usage *CursorUsage `json:"usage,omitempty"`
}

// CursorUsage Cursor使用统计
type CursorUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// SSEEvent 服务器发送事件
type SSEEvent struct {
	Data  string `json:"data"`
	Event string `json:"event,omitempty"`
	ID    string `json:"id,omitempty"`
}

// GetStringContent 获取消息的字符串内容
func (m *Message) GetStringContent() string {
	if m.Content == nil {
		return ""
	}

	switch content := m.Content.(type) {
	case string:
		return content
	case []ContentPart:
		var text string
		for _, part := range content {
			if part.Type == "text" {
				text += part.Text
			}
		}
		return text
	case []interface{}:
		// 处理混合类型内容
		var text string
		for _, item := range content {
			if part, ok := item.(map[string]interface{}); ok {
				if partType, exists := part["type"].(string); exists && partType == "text" {
					if textContent, exists := part["text"].(string); exists {
						text += textContent
					}
				}
			}
		}
		return text
	default:
		// 尝试将其他类型转换为JSON字符串
		if data, err := json.Marshal(content); err == nil {
			return string(data)
		}
		return ""
	}
}

// ToCursorMessages 将OpenAI消息转换为Cursor格式
// 注意：Cursor API 要求对话必须以用户消息开始，所以系统消息会被合并到第一条用户消息中
func ToCursorMessages(messages []Message, systemPromptInject string) []CursorMessage {
	var result []CursorMessage
	var systemContent string
	
	// 收集系统提示内容
	if len(messages) > 0 && messages[0].Role == "system" {
		systemContent = messages[0].GetStringContent()
		messages = messages[1:] // 跳过系统消息
	}
	
	// 添加注入的系统提示
	if systemPromptInject != "" {
		if systemContent != "" {
			systemContent += "\n" + systemPromptInject
		} else {
			systemContent = systemPromptInject
		}
	}

	// 转换其余消息
	firstUserFound := false
	for _, msg := range messages {
		if msg.Role == "" {
			continue // 跳过空消息
		}

		msgContent := msg.GetStringContent()
		
		// 如果有系统内容，将其作为上下文添加到第一条用户消息前面
		// 不使用明显的标签，避免模型重复回答
		if !firstUserFound && msg.Role == "user" && systemContent != "" {
			// 系统内容作为隐式上下文，用户消息紧随其后
			msgContent = systemContent + "\n\n---\n\n" + msgContent
			firstUserFound = true
		} else if msg.Role == "user" {
			firstUserFound = true
		}

		cursorMsg := CursorMessage{
			Role: msg.Role,
			Parts: []CursorPart{
				{
					Type: "text",
					Text: msgContent,
				},
			},
		}
		result = append(result, cursorMsg)
	}
	
	// 如果没有用户消息但有系统内容，创建一个包含系统内容的用户消息
	if len(result) == 0 && systemContent != "" {
		result = append(result, CursorMessage{
			Role: "user",
			Parts: []CursorPart{
				{Type: "text", Text: systemContent},
			},
		})
	}

	return result
}

// NewChatCompletionResponse 创建聊天完成响应
func NewChatCompletionResponse(id, model, content string, usage Usage) *ChatCompletionResponse {
	return &ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: usage,
	}
}

// NewChatCompletionStreamResponse 创建流式响应
func NewChatCompletionStreamResponse(id, model, content string, finishReason *string) *ChatCompletionStreamResponse {
	return &ChatCompletionStreamResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: StreamDelta{
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message, errorType, code string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    errorType,
			Code:    code,
		},
	}
}