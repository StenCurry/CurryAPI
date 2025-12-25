package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ClaudeMessageRequest Claude API消息请求格式
type ClaudeMessageRequest struct {
	Model         string           `json:"model" binding:"required"`
	Messages      []ClaudeMessage  `json:"messages" binding:"required"`
	MaxTokens     int              `json:"max_tokens"` // 可选，默认值将在验证时设置
	Temperature   *float64         `json:"temperature,omitempty"`
	TopP          *float64         `json:"top_p,omitempty"`
	TopK          *int             `json:"top_k,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`
	System        interface{}      `json:"system,omitempty"` // 支持 string 或 []ClaudeContentBlock
	Metadata      *ClaudeMetadata  `json:"metadata,omitempty"`
	Tools         []ClaudeTool     `json:"tools,omitempty"`         // 工具定义
	ToolChoice    interface{}      `json:"tool_choice,omitempty"`   // 工具选择策略
}

// ClaudeTool Claude工具定义
type ClaudeTool struct {
	Type          string                 `json:"type"`                     // 工具类型: "custom", "text_editor_20250728", "bash_20250124" 等
	Name          string                 `json:"name,omitempty"`           // 工具名称
	Description   string                 `json:"description,omitempty"`   // 工具描述
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`  // 输入参数schema
	MaxCharacters int                    `json:"max_characters,omitempty"` // text_editor 专用参数
}

// ClaudeToolUse Claude工具使用请求
type ClaudeToolUse struct {
	Type  string                 `json:"type"` // "tool_use"
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ClaudeToolResult Claude工具结果
type ClaudeToolResult struct {
	Type      string      `json:"type"` // "tool_result"
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content"` // string 或 []ClaudeContentBlock
	IsError   bool        `json:"is_error,omitempty"`
}

// ClaudeMessage Claude消息格式
type ClaudeMessage struct {
	Role    string                 `json:"role" binding:"required"`
	Content interface{}            `json:"content" binding:"required"`
}

// ClaudeContentBlock Claude内容块
type ClaudeContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Source    *ClaudeImageSource     `json:"source,omitempty"`
	// Tool use fields
	ID        string                 `json:"id,omitempty"`    // tool_use ID
	Name      string                 `json:"name,omitempty"`  // tool name
	Input     map[string]interface{} `json:"input,omitempty"` // tool input
	// Tool result fields
	ToolUseID string                 `json:"tool_use_id,omitempty"` // for tool_result
	Content   interface{}            `json:"content,omitempty"`     // tool result content (can be string or nested blocks)
	IsError   bool                   `json:"is_error,omitempty"`    // for tool_result errors
}

// ClaudeImageSource Claude图片源
type ClaudeImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ClaudeMetadata Claude元数据
type ClaudeMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// ClaudeMessageResponse Claude消息响应格式
type ClaudeMessageResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Role         string                `json:"role"`
	Content      []ClaudeContentBlock  `json:"content"`
	Model        string                `json:"model"`
	StopReason   string                `json:"stop_reason,omitempty"` // "end_turn", "max_tokens", "stop_sequence", "tool_use"
	StopSequence *string               `json:"stop_sequence,omitempty"`
	Usage        ClaudeUsage           `json:"usage"`
}

// ClaudeStreamResponse Claude流式响应
type ClaudeStreamResponse struct {
	Type         string                `json:"type"`
	Index        int                   `json:"index,omitempty"`
	Delta        *ClaudeStreamDelta    `json:"delta,omitempty"`
	Message      *ClaudeMessageResponse `json:"message,omitempty"`
	ContentBlock *ClaudeContentBlock   `json:"content_block,omitempty"`
	Usage        *ClaudeUsage          `json:"usage,omitempty"`
}

// ClaudeStreamDelta Claude流式增量
type ClaudeStreamDelta struct {
	Type         string `json:"type,omitempty"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence"` // 使用指针以便输出null
}

// ClaudeUsage Claude使用统计
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeErrorResponse Claude错误响应
type ClaudeErrorResponse struct {
	Type  string            `json:"type"`
	Error ClaudeErrorDetail `json:"error"`
}

// ClaudeErrorDetail Claude错误详情
type ClaudeErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ToOpenAIRequest 将Claude请求转换为OpenAI格式
func (r *ClaudeMessageRequest) ToOpenAIRequest() *ChatCompletionRequest {
	openAIMessages := make([]Message, 0, len(r.Messages)+1)
	
	// 处理system参数 - 支持字符串或内容块数组
	if r.System != nil {
		systemContent := ""
		
		switch sys := r.System.(type) {
		case string:
			// 简单字符串格式
			systemContent = sys
		case []interface{}:
			// 数组格式 - 提取所有文本内容
			for _, item := range sys {
				if block, ok := item.(map[string]interface{}); ok {
					if blockType, exists := block["type"].(string); exists && blockType == "text" {
						if text, exists := block["text"].(string); exists {
							if systemContent != "" {
								systemContent += "\n"
							}
							systemContent += text
						}
					}
				}
			}
		case []ClaudeContentBlock:
			// 已解析的内容块数组
			for _, block := range sys {
				if block.Type == "text" {
					if systemContent != "" {
						systemContent += "\n"
					}
					systemContent += block.Text
				}
			}
		}
		
		// 如果提取到了系统内容，添加为第一条消息
		if systemContent != "" {
			openAIMessages = append(openAIMessages, Message{
				Role:    "system",
				Content: systemContent,
			})
		}
	}
	
	// 转换消息
	for _, msg := range r.Messages {
		openAIMsg := Message{
			Role: msg.Role,
		}
		
		// 处理content - 支持字符串和内容块数组
		switch content := msg.Content.(type) {
		case string:
			// 简单字符串内容
			openAIMsg.Content = content
		case []interface{}:
			// 处理多模态内容块数组
			var textParts []string
			for _, item := range content {
				if block, ok := item.(map[string]interface{}); ok {
					blockType, _ := block["type"].(string)
					
					switch blockType {
					case "text":
						if text, exists := block["text"].(string); exists && text != "" {
							textParts = append(textParts, text)
						}
					case "tool_result":
						// 处理工具结果 - 这是 Claude Code CLI 发送的工具执行结果
						// 使用简洁的格式，直接展示结果内容
						isError, _ := block["is_error"].(bool)
						
						var resultContent string
						switch c := block["content"].(type) {
						case string:
							resultContent = c
						case []interface{}:
							// 处理嵌套的内容块
							for _, nested := range c {
								if nestedBlock, ok := nested.(map[string]interface{}); ok {
									if nestedBlock["type"] == "text" {
										if text, exists := nestedBlock["text"].(string); exists {
											resultContent += text
										}
									}
								}
							}
						}
						
						// 简化格式：直接展示工具执行结果
						// 不使用复杂的标签，避免模型混淆
						if isError {
							textParts = append(textParts, fmt.Sprintf("Tool execution failed:\n%s", resultContent))
						} else {
							// 直接使用结果内容，不添加额外标签
							textParts = append(textParts, resultContent)
						}
					case "tool_use":
						// 处理工具调用（assistant 消息中的）
						// 这是之前 assistant 发起的工具调用，简化格式
						toolName, _ := block["name"].(string)
						toolInput, _ := block["input"].(map[string]interface{})
						inputJSON, _ := json.Marshal(toolInput)
						textParts = append(textParts, fmt.Sprintf("Used tool %s with input: %s", toolName, string(inputJSON)))
					}
					// 注意: 图片类型暂时忽略，因为当前后端不支持
				}
			}
			openAIMsg.Content = strings.Join(textParts, "\n\n")
		case []ClaudeContentBlock:
			// 处理已解析的内容块数组
			var textParts []string
			for _, block := range content {
				switch block.Type {
				case "text":
					if block.Text != "" {
						textParts = append(textParts, block.Text)
					}
				case "tool_result":
					var resultContent string
					switch c := block.Content.(type) {
					case string:
						resultContent = c
					}
					// 简化格式
					if block.IsError {
						textParts = append(textParts, fmt.Sprintf("Tool execution failed:\n%s", resultContent))
					} else {
						textParts = append(textParts, resultContent)
					}
				case "tool_use":
					inputJSON, _ := json.Marshal(block.Input)
					textParts = append(textParts, fmt.Sprintf("Used tool %s with input: %s", block.Name, string(inputJSON)))
				}
			}
			openAIMsg.Content = strings.Join(textParts, "\n\n")
		default:
			openAIMsg.Content = ""
		}
		
		openAIMessages = append(openAIMessages, openAIMsg)
	}
	
	// 构建OpenAI请求
	req := &ChatCompletionRequest{
		Model:       r.Model,
		Messages:    openAIMessages,
		Stream:      r.Stream,
		Temperature: r.Temperature,
		MaxTokens:   &r.MaxTokens,
		TopP:        r.TopP,
	}
	
	// 处理stop_sequences参数（映射到OpenAI的stop参数）
	if len(r.StopSequences) > 0 {
		req.Stop = r.StopSequences
	}
	
	// 处理metadata中的user_id（映射到OpenAI的user参数）
	if r.Metadata != nil && r.Metadata.UserID != "" {
		req.User = r.Metadata.UserID
	}
	
	// 处理tools参数（转换为OpenAI格式）
	if len(r.Tools) > 0 {
		openAITools := make([]Tool, 0, len(r.Tools))
		for _, tool := range r.Tools {
			// 处理 Anthropic 定义的工具类型（如 text_editor_20250728）
			if tool.Type != "" && tool.Type != "custom" && tool.Name == "" {
				// 这是 Anthropic 内置工具，如 text_editor_20250728
				// 转换为 OpenAI function 格式
				openAITool := Tool{
					Type: "function",
					Function: &FunctionDefinition{
						Name:        tool.Type, // 使用 type 作为 name
						Description: "Anthropic built-in tool: " + tool.Type,
						Parameters:  tool.InputSchema,
					},
				}
				openAITools = append(openAITools, openAITool)
			} else if tool.Name != "" {
				// 自定义工具
				openAITool := Tool{
					Type: "function",
					Function: &FunctionDefinition{
						Name:        tool.Name,
						Description: tool.Description,
						Parameters:  tool.InputSchema,
					},
				}
				openAITools = append(openAITools, openAITool)
			}
		}
		if len(openAITools) > 0 {
			req.Tools = openAITools
		}
	}
	
	return req
}

// NewClaudeMessageResponse 从OpenAI响应创建Claude响应
func NewClaudeMessageResponse(openAIResp *ChatCompletionResponse) *ClaudeMessageResponse {
	contentBlocks := []ClaudeContentBlock{}
	finishReason := "end_turn"
	
	if len(openAIResp.Choices) > 0 {
		choice := openAIResp.Choices[0]
		
		// 添加文本内容（如果有）
		textContent := choice.Message.GetStringContent()
		if textContent != "" {
			contentBlocks = append(contentBlocks, ClaudeContentBlock{
				Type: "text",
				Text: textContent,
			})
		}
		
		// 处理工具调用
		if len(choice.Message.ToolCalls) > 0 {
			finishReason = "tool_use"
			for _, toolCall := range choice.Message.ToolCalls {
				// 解析工具参数
				var input map[string]interface{}
				if toolCall.Function.Arguments != "" {
					// 尝试解析 JSON 参数
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
						// 如果解析失败，将整个参数作为字符串
						input = map[string]interface{}{"raw": toolCall.Function.Arguments}
					}
				}
				
				contentBlocks = append(contentBlocks, ClaudeContentBlock{
					Type:  "tool_use",
					ID:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: input,
				})
			}
		}
		
		// 映射finish_reason到Claude的stop_reason
		switch choice.FinishReason {
		case "stop":
			if finishReason != "tool_use" {
				finishReason = "end_turn"
			}
		case "length":
			finishReason = "max_tokens"
		case "content_filter":
			finishReason = "stop_sequence"
		case "tool_calls", "function_call":
			finishReason = "tool_use"
		default:
			if finishReason != "tool_use" {
				finishReason = "end_turn"
			}
		}
	}
	
	// 如果没有任何内容块，添加一个空文本块
	if len(contentBlocks) == 0 {
		contentBlocks = append(contentBlocks, ClaudeContentBlock{
			Type: "text",
			Text: "",
		})
	}
	
	return &ClaudeMessageResponse{
		ID:         openAIResp.ID,
		Type:       "message",
		Role:       "assistant",
		Content:    contentBlocks,
		Model:      openAIResp.Model,
		StopReason: finishReason,
		Usage: ClaudeUsage{
			InputTokens:  openAIResp.Usage.PromptTokens,
			OutputTokens: openAIResp.Usage.CompletionTokens,
		},
	}
}

// NewClaudeStreamResponse 创建Claude流式响应
// 参数:
//   - eventType: 事件类型 (message_start, content_block_start, content_block_delta, content_block_stop, message_delta, message_stop)
//   - text: 文本内容（用于content_block_delta）
//   - stopReason: 停止原因（用于message_delta）
//   - model: 模型名称（用于message_start）
//   - messageID: 消息ID（用于message_start）
//   - usage: 使用统计（用于message_start和message_delta）
func NewClaudeStreamResponse(eventType string, text string, stopReason string) *ClaudeStreamResponse {
	resp := &ClaudeStreamResponse{
		Type: eventType,
	}
	
	switch eventType {
	case "message_start":
		resp.Message = &ClaudeMessageResponse{
			ID:      "msg-" + time.Now().Format("20060102150405"),
			Type:    "message",
			Role:    "assistant",
			Content: []ClaudeContentBlock{},
			Model:   "",
			Usage: ClaudeUsage{
				InputTokens:  0,
				OutputTokens: 0,
			},
		}
	case "content_block_start":
		resp.Index = 0
		resp.ContentBlock = &ClaudeContentBlock{
			Type: "text",
			Text: "",
		}
	case "content_block_delta":
		resp.Index = 0
		resp.Delta = &ClaudeStreamDelta{
			Type: "text_delta",
			Text: text,
		}
	case "content_block_stop":
		resp.Index = 0
	case "message_delta":
		resp.Delta = &ClaudeStreamDelta{
			StopReason: stopReason,
		}
		resp.Usage = &ClaudeUsage{
			OutputTokens: 0,
		}
	case "message_stop":
		// 空响应，仅包含type字段
	}
	
	return resp
}

// NewClaudeStreamResponseWithDetails 创建带详细信息的Claude流式响应
func NewClaudeStreamResponseWithDetails(eventType, text, stopReason, model, messageID string, inputTokens, outputTokens int) *ClaudeStreamResponse {
	resp := &ClaudeStreamResponse{
		Type: eventType,
	}
	
	switch eventType {
	case "message_start":
		resp.Message = &ClaudeMessageResponse{
			ID:      messageID,
			Type:    "message",
			Role:    "assistant",
			Content: []ClaudeContentBlock{},
			Model:   model,
			Usage: ClaudeUsage{
				InputTokens:  inputTokens,
				OutputTokens: 0,
			},
		}
	case "content_block_start":
		resp.Index = 0
		resp.ContentBlock = &ClaudeContentBlock{
			Type: "text",
			Text: "",
		}
	case "content_block_delta":
		resp.Index = 0
		resp.Delta = &ClaudeStreamDelta{
			Type: "text_delta",
			Text: text,
		}
	case "content_block_stop":
		resp.Index = 0
	case "message_delta":
		delta := &ClaudeStreamDelta{}
		if stopReason != "" {
			delta.StopReason = stopReason
		}
		resp.Delta = delta
		
		resp.Usage = &ClaudeUsage{
			OutputTokens: outputTokens,
		}
	case "message_stop":
		// 空响应，仅包含type字段
	}
	
	return resp
}

// MapOpenAIFinishReasonToClaude 将OpenAI的finish_reason映射到Claude的stop_reason
func MapOpenAIFinishReasonToClaude(finishReason string) string {
	switch finishReason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "content_filter":
		return "stop_sequence"
	case "tool_calls", "function_call":
		return "tool_use" // 工具调用时返回 tool_use
	default:
		return "end_turn"
	}
}

// NewClaudeErrorResponse 创建Claude错误响应
func NewClaudeErrorResponse(errorType, message string) *ClaudeErrorResponse {
	return &ClaudeErrorResponse{
		Type: "error",
		Error: ClaudeErrorDetail{
			Type:    errorType,
			Message: message,
		},
	}
}

// NewClaudeInvalidRequestError 创建无效请求错误
func NewClaudeInvalidRequestError(message string) *ClaudeErrorResponse {
	return NewClaudeErrorResponse("invalid_request_error", message)
}

// NewClaudeAuthenticationError 创建认证错误
func NewClaudeAuthenticationError(message string) *ClaudeErrorResponse {
	if message == "" {
		message = "Invalid API key"
	}
	return NewClaudeErrorResponse("authentication_error", message)
}

// NewClaudeRateLimitError 创建速率限制错误
func NewClaudeRateLimitError(message string) *ClaudeErrorResponse {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return NewClaudeErrorResponse("rate_limit_error", message)
}

// NewClaudeAPIError 创建API错误
func NewClaudeAPIError(message string) *ClaudeErrorResponse {
	if message == "" {
		message = "Internal server error"
	}
	return NewClaudeErrorResponse("api_error", message)
}

// NewClaudeOverloadedError 创建过载错误
func NewClaudeOverloadedError(message string) *ClaudeErrorResponse {
	if message == "" {
		message = "Service is currently overloaded"
	}
	return NewClaudeErrorResponse("overloaded_error", message)
}

// Validate 验证Claude请求的有效性
func (r *ClaudeMessageRequest) Validate() error {
	// 验证必需字段
	if r.Model == "" {
		return &ValidationError{Field: "model", Message: "model is required"}
	}
	
	if len(r.Messages) == 0 {
		return &ValidationError{Field: "messages", Message: "messages array cannot be empty"}
	}
	
	if r.MaxTokens <= 0 {
		return &ValidationError{Field: "max_tokens", Message: "max_tokens must be greater than 0"}
	}
	
	// 验证消息格式
	for i, msg := range r.Messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			return &ValidationError{
				Field:   "messages",
				Message: "message role must be 'user' or 'assistant' at index " + string(rune(i)),
			}
		}
		
		if msg.Content == nil {
			return &ValidationError{
				Field:   "messages",
				Message: "message content cannot be null at index " + string(rune(i)),
			}
		}
	}
	
	// 验证参数范围
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 1) {
		return &ValidationError{Field: "temperature", Message: "temperature must be between 0 and 1"}
	}
	
	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return &ValidationError{Field: "top_p", Message: "top_p must be between 0 and 1"}
	}
	
	if r.TopK != nil && *r.TopK < 0 {
		return &ValidationError{Field: "top_k", Message: "top_k must be non-negative"}
	}
	
	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
