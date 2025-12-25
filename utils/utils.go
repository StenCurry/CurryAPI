package utils

import (
	"bufio"
	"context"
	"crypto/rand"
	"Curry2API-go/models"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	byteLen := (length + 1) / 2
	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		fallback := fmt.Sprintf("%d", time.Now().UnixNano())
		if len(fallback) >= length {
			return fallback[:length]
		}
		return fallback
	}

	encoded := hex.EncodeToString(bytes)
	if len(encoded) < length {
		encoded += GenerateRandomString(length - len(encoded))
	}

	return encoded[:length]
}

// GenerateChatCompletionID 生成聊天完成ID
func GenerateChatCompletionID() string {
	return "chatcmpl-" + GenerateRandomString(29)
}

// ParseSSELine 解析SSE数据行
func ParseSSELine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "data: ") {
		return strings.TrimSpace(line[6:]) // 去掉 'data: ' 前缀并去除前导空格
	}
	return ""
}

// WriteSSEEvent 写入SSE事件
// 优化：立即刷新缓冲区以减少延迟
func WriteSSEEvent(w http.ResponseWriter, event, data string) error {
	// 使用单次写入减少系统调用
	var buf strings.Builder
	if event != "" {
		buf.WriteString("event: ")
		buf.WriteString(event)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	buf.WriteString(data)
	buf.WriteString("\n\n")
	
	if _, err := w.Write([]byte(buf.String())); err != nil {
		return err
	}

	// 立即刷新缓冲区以减少延迟
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// StreamChatCompletion 处理流式聊天完成
func StreamChatCompletion(c *gin.Context, chatGenerator <-chan interface{}) {
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

	// 生成响应ID
	responseID := GenerateChatCompletionID()
	
	// Track usage data as we stream
	var accumulatedUsage models.Usage
	var streamError error

	// 处理流式数据
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			logrus.Debug("Client disconnected during streaming")
			// Track incomplete request if tracking function is available
			if trackFunc, exists := c.Get("track_usage_func"); exists {
				if fn, ok := trackFunc.(UsageTrackingFunc); ok {
					fn(c, nil, 499, "Client disconnected")
				}
			}
			return

		case data, ok := <-chatGenerator:
			if !ok {
				// 通道关闭，发送完成事件
				finishEvent := models.NewChatCompletionStreamResponse(responseID, "gpt-4o", "", stringPtr("stop"))
				if jsonData, err := json.Marshal(finishEvent); err == nil {
					WriteSSEEvent(c.Writer, "", string(jsonData))
				}
				WriteSSEEvent(c.Writer, "", "[DONE]")
				
				// Track successful streaming request if tracking function is available
				if streamError == nil {
					if trackFunc, exists := c.Get("track_usage_func"); exists {
						if fn, ok := trackFunc.(UsageTrackingFunc); ok {
							fn(c, &accumulatedUsage, http.StatusOK, "")
						}
					}
				}
				return
			}

			switch v := data.(type) {
			case string:
				// 文本内容
				if v != "" {
					streamResp := models.NewChatCompletionStreamResponse(responseID, "gpt-4o", v, nil)
					if jsonData, err := json.Marshal(streamResp); err == nil {
						WriteSSEEvent(c.Writer, "", string(jsonData))
					}
				}

			case models.Usage:
				// 使用统计 - 累积token使用情况
				accumulatedUsage.PromptTokens += v.PromptTokens
				accumulatedUsage.CompletionTokens += v.CompletionTokens
				accumulatedUsage.TotalTokens += v.TotalTokens
				continue

			case error:
				streamError = v
				logrus.WithError(v).Error("Stream generator error")
				WriteSSEEvent(c.Writer, "", "[DONE]")
				// Track failed streaming request if tracking function is available
				if trackFunc, exists := c.Get("track_usage_func"); exists {
					if fn, ok := trackFunc.(UsageTrackingFunc); ok {
						fn(c, nil, http.StatusInternalServerError, v.Error())
					}
				}
				return

			default:
				logrus.Warnf("Unknown data type in stream: %T", v)
			}
		}
	}
}

// UsageTrackingFunc is a function type for tracking usage
type UsageTrackingFunc func(c *gin.Context, usage *models.Usage, statusCode int, errorMsg string)

// NonStreamChatCompletion 处理非流式聊天完成
func NonStreamChatCompletion(c *gin.Context, chatGenerator <-chan interface{}) {
	var fullContent strings.Builder
	var usage models.Usage
	var streamError error

	// 收集所有数据
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, models.NewErrorResponse(
				"Request timeout",
				"timeout_error",
				"request_timeout",
			))
			// Track failed request if tracking function is available
			if trackFunc, exists := c.Get("track_usage_func"); exists {
				if fn, ok := trackFunc.(UsageTrackingFunc); ok {
					fn(c, nil, http.StatusRequestTimeout, "Request timeout")
				}
			}
			return

		case data, ok := <-chatGenerator:
			if !ok {
				// 数据收集完成，返回响应
				responseID := GenerateChatCompletionID()
				response := models.NewChatCompletionResponse(
					responseID,
					"gpt-4o",
					fullContent.String(),
					usage,
				)
				
				// Track successful request with usage data if tracking function is available
				if streamError == nil {
					if trackFunc, exists := c.Get("track_usage_func"); exists {
						if fn, ok := trackFunc.(UsageTrackingFunc); ok {
							fn(c, &usage, http.StatusOK, "")
						}
					}
				}
				
				c.JSON(http.StatusOK, response)
				return
			}

			switch v := data.(type) {
			case string:
				fullContent.WriteString(v)
			case models.Usage:
				usage = v
			case error:
				streamError = v
				logrus.WithError(v).Error("Stream generator error")
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
					"Internal server error",
					"stream_error",
					"",
				))
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

// ErrorWrapper 错误包装器
func ErrorWrapper(handler func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			logrus.WithError(err).Error("Handler error")

			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
					"Internal server error",
					"internal_error",
					"",
				))
			}
		}
	}
}

// SafeStreamWrapper 安全流式包装器
func SafeStreamWrapper(handler func(*gin.Context, <-chan interface{}), c *gin.Context, chatGenerator <-chan interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("Panic in stream handler")
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
					"Internal server error",
					"panic_error",
					"",
				))
			}
		}
	}()

	firstItem, ok := <-chatGenerator
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Empty stream",
			"empty_stream",
			"",
		))
		return
	}

	if err, isErr := firstItem.(error); isErr {
		logrus.WithError(err).Error("Stream error")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			err.Error(),
			"stream_error",
			"",
		))
		return
	}

	buffered := make(chan interface{}, 1)
	buffered <- firstItem
	ctx := c.Request.Context()

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

	handler(c, buffered)
}

// CreateHTTPClient 创建HTTP客户端
func CreateHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// ReadSSEStream 读取SSE流
// 优化：使用更大的缓冲区和更高效的读取方式
func ReadSSEStream(ctx context.Context, resp *http.Response, output chan<- interface{}) error {
	// 使用更大的缓冲区以提高读取效率
	reader := bufio.NewReaderSize(resp.Body, 128*1024) // 128KB 缓冲区
	defer resp.Body.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		data := ParseSSELine(line)
		if data == "" {
			continue
		}

		if data == "[DONE]" {
			return nil
		}

		// 尝试解析JSON数据
		var eventData models.CursorEventData
		if err := json.Unmarshal([]byte(data), &eventData); err != nil {
			logrus.WithError(err).Debugf("Failed to parse SSE data: %s", data)
			continue
		}

		// 处理不同类型的事件
		switch eventData.Type {
		case "error":
			if eventData.ErrorText != "" {
				return fmt.Errorf("cursor API error: %s", eventData.ErrorText)
			}

		case "finish":
			if eventData.MessageMetadata != nil && eventData.MessageMetadata.Usage != nil {
				usage := models.Usage{
					PromptTokens:     eventData.MessageMetadata.Usage.InputTokens,
					CompletionTokens: eventData.MessageMetadata.Usage.OutputTokens,
					TotalTokens:      eventData.MessageMetadata.Usage.TotalTokens,
				}
				select {
				case output <- usage:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			// 不要在这里返回！继续读取可能还有更多内容
			// Cursor API 可能在长回答中发送多个 finish 事件
			continue

		default:
			if eventData.Delta != "" {
				select {
				case output <- eventData.Delta:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// ValidateModel 验证模型名称
func ValidateModel(model string, validModels []string) bool {
	for _, validModel := range validModels {
		if validModel == model {
			return true
		}
	}
	return false
}

// SanitizeContent 清理内容
func SanitizeContent(content string) string {
	// 移除可能的恶意内容
	content = strings.ReplaceAll(content, "\x00", "")
	return content
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// CopyHeaders 复制HTTP头
func CopyHeaders(dst, src http.Header, skipHeaders []string) {
	skipMap := make(map[string]bool)
	for _, header := range skipHeaders {
		skipMap[strings.ToLower(header)] = true
	}

	for key, values := range src {
		if skipMap[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// IsJSONContentType 检查是否为JSON内容类型
func IsJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// ReadRequestBody 读取请求体
func ReadRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	return body, nil
}

// RunJS 执行JavaScript代码并返回标准输出内容
func RunJS(jsCode string) (string, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "cursor_js_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建JavaScript文件
	jsFilePath := filepath.Join(tempDir, "script.js")
	file, err := os.Create(jsFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create js file: %w", err)
	}

	// 添加crypto模块导入并设置为全局变量
	jsContent := `const crypto = require('crypto').webcrypto;
global.crypto = crypto;
globalThis.crypto = crypto;
// 在Node.js环境中创建window对象
if (typeof window === 'undefined') { global.window = global; }
window.crypto = crypto;
this.crypto = crypto;
` + jsCode

	if _, err := file.WriteString(jsContent); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to write js content: %w", err)
	}
	file.Close()

	// 执行Node.js命令
	cmd := exec.Command("node", jsFilePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("node.js execution failed (exit code: %d)\nSTDOUT:\n%s\nSTDERR:\n%s",
				exitErr.ExitCode(), string(output), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute node.js: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
