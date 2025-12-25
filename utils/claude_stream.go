package utils

import (
	"Curry2API-go/models"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ParseToolCallFromContent 从内容中解析工具调用
// 返回: 工具调用对象, 工具调用之前的文本, 是否找到工具调用
func ParseToolCallFromContent(content string) (*models.ClaudeToolUse, string, bool) {
	// 匹配 <tool_call>...</tool_call> 格式
	toolCallRegex := regexp.MustCompile(`(?s)<tool_call>\s*<tool_name>([^<]+)</tool_name>\s*<tool_input>\s*(.*?)\s*</tool_input>\s*</tool_call>`)
	
	matches := toolCallRegex.FindStringSubmatch(content)
	if len(matches) < 3 {
		return nil, content, false
	}
	
	toolName := strings.TrimSpace(matches[1])
	toolInputStr := strings.TrimSpace(matches[2])
	
	// 解析 JSON 输入
	var toolInput map[string]interface{}
	if err := json.Unmarshal([]byte(toolInputStr), &toolInput); err != nil {
		logrus.WithError(err).Warn("Failed to parse tool input JSON, trying to clean")
		// 尝试清理 JSON
		toolInputStr = strings.ReplaceAll(toolInputStr, "\n", "")
		toolInputStr = strings.TrimSpace(toolInputStr)
		if err := json.Unmarshal([]byte(toolInputStr), &toolInput); err != nil {
			logrus.WithError(err).Error("Failed to parse tool input JSON after cleaning")
			return nil, content, false
		}
	}
	
	// 生成工具调用 ID
	toolUseID := fmt.Sprintf("toolu_%s", GenerateRandomString(24))
	
	toolUse := &models.ClaudeToolUse{
		Type:  "tool_use",
		ID:    toolUseID,
		Name:  toolName,
		Input: toolInput,
	}
	
	// 提取工具调用之前的文本
	idx := strings.Index(content, "<tool_call>")
	beforeText := ""
	if idx > 0 {
		beforeText = strings.TrimSpace(content[:idx])
	}
	
	return toolUse, beforeText, true
}

// StreamClaudeCompletion 处理Claude流式响应
// 按照Claude API规范发送SSE事件序列:
// 1. message_start - 消息开始
// 2. content_block_start - 内容块开始
// 3. content_block_delta - 内容增量（多次）
// 4. content_block_stop - 内容块结束
// 5. message_delta - 消息元数据（包含stop_reason和usage）
// 6. message_stop - 消息结束
func StreamClaudeCompletion(c *gin.Context, chatGenerator <-chan interface{}) {
	// 设置SSE头 - 关键配置以确保流式响应立即发送
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	// 禁用Nginx/代理缓冲 - 这是关键！
	c.Header("X-Accel-Buffering", "no")
	// 禁用压缩以避免缓冲
	c.Header("Content-Encoding", "identity")
	// 设置Transfer-Encoding为chunked
	c.Header("Transfer-Encoding", "chunked")
	
	// 立即刷新头部
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 生成消息ID
	messageID := "msg-" + GenerateRandomString(29)
	model := "claude-3-5-sonnet-20241022" // 默认模型

	// 发送 message_start 事件
	messageStartEvent := models.NewClaudeStreamResponseWithDetails(
		"message_start",
		"",
		"",
		model,
		messageID,
		0, // input_tokens 将在后续更新
		0,
	)
	if err := writeClaudeSSEEvent(c.Writer, messageStartEvent); err != nil {
		logrus.WithError(err).Error("Failed to write message_start event")
		return
	}

	// 发送 content_block_start 事件
	contentBlockStartEvent := models.NewClaudeStreamResponse("content_block_start", "", "")
	if err := writeClaudeSSEEvent(c.Writer, contentBlockStartEvent); err != nil {
		logrus.WithError(err).Error("Failed to write content_block_start event")
		return
	}

	// 处理流式数据
	ctx := c.Request.Context()
	var usage models.Usage
	hasContent := false
	stopReason := "end_turn"
	
	// 检查是否需要解析工具调用
	hasToolUse, _ := c.Get("has_tool_use")
	var contentBuffer strings.Builder
	toolCallDetected := false

	for {
		select {
		case <-ctx.Done():
			logrus.Debug("Client disconnected during Claude streaming")
			// Track incomplete request if tracking function is available
			if trackFunc, exists := c.Get("track_usage_func"); exists {
				if fn, ok := trackFunc.(UsageTrackingFunc); ok {
					fn(c, nil, 499, "Client disconnected")
				}
			}
			return

		case data, ok := <-chatGenerator:
			if !ok {
				// 通道关闭
				
				// 如果启用了工具调用检测，检查缓冲区中是否有工具调用
				if hasToolUse == true {
					fullContent := contentBuffer.String()
					toolUse, _, found := ParseToolCallFromContent(fullContent)
					if found {
						toolCallDetected = true
						stopReason = "tool_use"
						
						// 发送 content_block_stop 事件（结束文本块）
						contentBlockStopEvent := models.NewClaudeStreamResponse("content_block_stop", "", "")
						writeClaudeSSEEvent(c.Writer, contentBlockStopEvent)
						
						// 工具调用时不发送 beforeText，直接发送工具调用块
						// 这样可以避免重复回答的问题
						
						// 发送 content_block_start 事件（工具调用块）
						toolBlockStartEvent := &models.ClaudeStreamResponse{
							Type:  "content_block_start",
							Index: 1,
							ContentBlock: &models.ClaudeContentBlock{
								Type:  "tool_use",
								ID:    toolUse.ID,
								Name:  toolUse.Name,
								Input: map[string]interface{}{},
							},
						}
						writeClaudeSSEEvent(c.Writer, toolBlockStartEvent)
						
						// 发送 content_block_delta 事件（工具输入）
						inputJSON, _ := json.Marshal(toolUse.Input)
						toolDeltaEvent := &models.ClaudeStreamResponse{
							Type:  "content_block_delta",
							Index: 1,
							Delta: &models.ClaudeStreamDelta{
								Type: "input_json_delta",
								Text: string(inputJSON),
							},
						}
						writeClaudeSSEEvent(c.Writer, toolDeltaEvent)
						
						// 发送 content_block_stop 事件（工具调用块结束）
						toolBlockStopEvent := &models.ClaudeStreamResponse{
							Type:  "content_block_stop",
							Index: 1,
						}
						writeClaudeSSEEvent(c.Writer, toolBlockStopEvent)
					} else {
						// 没有工具调用，正常结束
						contentBlockStopEvent := models.NewClaudeStreamResponse("content_block_stop", "", "")
						writeClaudeSSEEvent(c.Writer, contentBlockStopEvent)
					}
				} else {
					// 发送 content_block_stop 事件
					contentBlockStopEvent := models.NewClaudeStreamResponse("content_block_stop", "", "")
					if err := writeClaudeSSEEvent(c.Writer, contentBlockStopEvent); err != nil {
						logrus.WithError(err).Error("Failed to write content_block_stop event")
					}
				}

				// 发送 message_delta 事件（包含stop_reason和usage）
				messageDeltaEvent := models.NewClaudeStreamResponseWithDetails(
					"message_delta",
					"",
					stopReason,
					"",
					"",
					0,
					usage.CompletionTokens,
				)
				if err := writeClaudeSSEEvent(c.Writer, messageDeltaEvent); err != nil {
					logrus.WithError(err).Error("Failed to write message_delta event")
				}

				// 发送 message_stop 事件
				messageStopEvent := models.NewClaudeStreamResponse("message_stop", "", "")
				if err := writeClaudeSSEEvent(c.Writer, messageStopEvent); err != nil {
					logrus.WithError(err).Error("Failed to write message_stop event")
				}

				// Track successful streaming request if tracking function is available
				if trackFunc, exists := c.Get("track_usage_func"); exists {
					if fn, ok := trackFunc.(UsageTrackingFunc); ok {
						fn(c, &usage, http.StatusOK, "")
					}
				}

				return
			}

			switch v := data.(type) {
			case string:
				// 文本内容 - 发送 content_block_delta 事件
				if v != "" {
					hasContent = true
					
					// 如果启用了工具调用检测，缓冲内容并智能发送
					if hasToolUse == true {
						contentBuffer.WriteString(v)
						fullContent := contentBuffer.String()
						
						// 检查是否包含工具调用标签开始
						if strings.Contains(fullContent, "<tool_call>") {
							// 检测到工具调用开始，停止发送文本
							// 不发送 beforeText，因为工具调用时 CLI 不需要额外的文本
							toolCallDetected = true
							continue
						}
						
						// 如果已经检测到工具调用开始，不发送后续内容
						if toolCallDetected {
							continue
						}
						
						// 没有检测到工具调用，正常发送
						// 但是要小心，不要发送可能是工具调用开头的内容
						// 检查是否可能是工具调用的开始（部分匹配）
						if strings.HasSuffix(fullContent, "<") || 
						   strings.HasSuffix(fullContent, "<t") ||
						   strings.HasSuffix(fullContent, "<to") ||
						   strings.HasSuffix(fullContent, "<too") ||
						   strings.HasSuffix(fullContent, "<tool") ||
						   strings.HasSuffix(fullContent, "<tool_") ||
						   strings.HasSuffix(fullContent, "<tool_c") ||
						   strings.HasSuffix(fullContent, "<tool_ca") ||
						   strings.HasSuffix(fullContent, "<tool_cal") ||
						   strings.HasSuffix(fullContent, "<tool_call") {
							// 可能是工具调用的开始，暂不发送，等待更多内容
							continue
						}
					}
					
					deltaEvent := models.NewClaudeStreamResponse("content_block_delta", v, "")
					if err := writeClaudeSSEEvent(c.Writer, deltaEvent); err != nil {
						logrus.WithError(err).Error("Failed to write content_block_delta event")
						return
					}
				}

			case models.Usage:
				// 使用统计 - 保存以便在message_delta中发送
				usage = v

			case error:
				logrus.WithError(v).Error("Stream generator error")
				
				// 发送错误事件
				errorResp := models.NewClaudeAPIError(v.Error())
				if jsonData, err := json.Marshal(errorResp); err == nil {
					WriteSSEEvent(c.Writer, "error", string(jsonData))
				}
				
				// 如果已经发送了内容，需要正常结束流
				if hasContent {
					contentBlockStopEvent := models.NewClaudeStreamResponse("content_block_stop", "", "")
					writeClaudeSSEEvent(c.Writer, contentBlockStopEvent)
					
					messageDeltaEvent := models.NewClaudeStreamResponseWithDetails(
						"message_delta",
						"",
						"error",
						"",
						"",
						0,
						usage.CompletionTokens,
					)
					writeClaudeSSEEvent(c.Writer, messageDeltaEvent)
					
					messageStopEvent := models.NewClaudeStreamResponse("message_stop", "", "")
					writeClaudeSSEEvent(c.Writer, messageStopEvent)
				}
				
				// Track failed streaming request if tracking function is available
				if trackFunc, exists := c.Get("track_usage_func"); exists {
					if fn, ok := trackFunc.(UsageTrackingFunc); ok {
						fn(c, nil, http.StatusInternalServerError, v.Error())
					}
				}
				return

			default:
				logrus.Warnf("Unknown data type in Claude stream: %T", v)
			}
		}
	}
}

// NonStreamClaudeCompletion 处理Claude非流式响应
// 收集所有数据后返回完整的Claude MessageResponse格式
// 支持工具调用解析
func NonStreamClaudeCompletion(c *gin.Context, chatGenerator <-chan interface{}) {
	var fullContent strings.Builder
	var usage models.Usage

	// 收集所有数据
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			// 请求超时
			errorResp := models.NewClaudeAPIError("Request timeout")
			c.JSON(http.StatusRequestTimeout, errorResp)
			// Track failed request if tracking function is available
			if trackFunc, exists := c.Get("track_usage_func"); exists {
				if fn, ok := trackFunc.(UsageTrackingFunc); ok {
					fn(c, nil, http.StatusRequestTimeout, "Request timeout")
				}
			}
			return

		case data, ok := <-chatGenerator:
			if !ok {
				// 数据收集完成，构建并返回Claude响应
				messageID := "msg-" + GenerateRandomString(29)
				model := "claude-3-5-sonnet-20241022" // 默认模型
				
				content := fullContent.String()
				stopReason := "end_turn"
				contentBlocks := []models.ClaudeContentBlock{}
				
				// 检查是否包含工具调用
				hasToolUse, _ := c.Get("has_tool_use")
				if hasToolUse == true {
					// 尝试解析工具调用
					toolUse, _, found := ParseToolCallFromContent(content)
					if found {
						// 工具调用时不添加 beforeText，直接返回工具调用块
						// 这样可以避免重复回答的问题
						contentBlocks = append(contentBlocks, models.ClaudeContentBlock{
							Type:  "tool_use",
							ID:    toolUse.ID,
							Name:  toolUse.Name,
							Input: toolUse.Input,
						})
						stopReason = "tool_use"
					} else {
						// 没有找到工具调用，使用原始内容
						contentBlocks = append(contentBlocks, models.ClaudeContentBlock{
							Type: "text",
							Text: content,
						})
					}
				} else {
					// 普通响应
					contentBlocks = append(contentBlocks, models.ClaudeContentBlock{
						Type: "text",
						Text: content,
					})
				}
				
				response := &models.ClaudeMessageResponse{
					ID:         messageID,
					Type:       "message",
					Role:       "assistant",
					Content:    contentBlocks,
					Model:      model,
					StopReason: stopReason,
					Usage: models.ClaudeUsage{
						InputTokens:  usage.PromptTokens,
						OutputTokens: usage.CompletionTokens,
					},
				}
				
				// Track successful request with usage data if tracking function is available
				if trackFunc, exists := c.Get("track_usage_func"); exists {
					if fn, ok := trackFunc.(UsageTrackingFunc); ok {
						fn(c, &usage, http.StatusOK, "")
					}
				}
				
				c.JSON(http.StatusOK, response)
				return
			}

			switch v := data.(type) {
			case string:
				// 累积文本内容
				fullContent.WriteString(v)
				
			case models.Usage:
				// 保存使用统计
				usage = v
				
			case error:
				// 错误处理
				logrus.WithError(v).Error("Error in Claude non-stream completion")
				errorResp := models.NewClaudeAPIError(v.Error())
				c.JSON(http.StatusInternalServerError, errorResp)
				// Track failed request if tracking function is available
				if trackFunc, exists := c.Get("track_usage_func"); exists {
					if fn, ok := trackFunc.(UsageTrackingFunc); ok {
						fn(c, nil, http.StatusInternalServerError, v.Error())
					}
				}
				return
			}
		}
	}
}

// writeClaudeSSEEvent 写入Claude SSE事件
// Claude API使用标准SSE格式，每个事件包含event和data字段
// 优化：使用单次写入和立即刷新以减少延迟
func writeClaudeSSEEvent(w http.ResponseWriter, event *models.ClaudeStreamResponse) error {
	// 序列化事件为JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 直接构建完整的SSE消息并一次性写入
	// 格式: event: <type>\ndata: <json>\n\n
	message := "event: " + event.Type + "\ndata: " + string(jsonData) + "\n\n"
	
	if _, err := w.Write([]byte(message)); err != nil {
		return err
	}

	// 立即刷新缓冲区以减少延迟 - 这是关键！
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// SafeClaudeStreamWrapper 安全的Claude流式包装器
// 处理panic和错误，确保资源正确清理
func SafeClaudeStreamWrapper(handler func(*gin.Context, <-chan interface{}), c *gin.Context, chatGenerator <-chan interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("Panic in Claude stream handler")
			if !c.Writer.Written() {
				errorResp := models.NewClaudeAPIError("Internal server error")
				c.JSON(http.StatusInternalServerError, errorResp)
			}
		}
	}()

	// 检查第一个项目以确定是否有错误
	firstItem, ok := <-chatGenerator
	if !ok {
		errorResp := models.NewClaudeAPIError("Empty stream")
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	// 如果第一个项目是错误，直接返回错误响应
	if err, isErr := firstItem.(error); isErr {
		logrus.WithError(err).Error("Error in first stream item")
		errorResp := models.NewClaudeAPIError(err.Error())
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	// 创建缓冲通道并重新放入第一个项目
	buffered := make(chan interface{}, 10)
	buffered <- firstItem
	ctx := c.Request.Context()

	// 启动goroutine转发剩余数据
	go func() {
		defer close(buffered)
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-chatGenerator:
				if !ok {
					return
				}
				select {
				case buffered <- item:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// 调用实际的处理器
	handler(c, buffered)
}
