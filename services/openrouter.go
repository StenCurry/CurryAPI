package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Curry2API-go/config"
	"Curry2API-go/models"
	"github.com/sirupsen/logrus"
)

// OpenRouterService OpenRouter API 服务
type OpenRouterService struct {
	config  *config.Config
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewOpenRouterService 创建新的 OpenRouter 服务
func NewOpenRouterService(cfg *config.Config) *OpenRouterService {
	return &OpenRouterService{
		config:  cfg,
		client:  &http.Client{Timeout: 120 * time.Second},
		apiKey:  "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",
		baseURL: "https://openrouter.ai/api/v1",
	}
}

// OpenRouter 免费模型列表
var openRouterFreeModels = map[string]bool{
	// Alibaba
	"alibaba/tongyi-deepresearch-30b-a3b": true,
	// AllenAI
	"allenai/olmo-3-32b-think": true,
	// Amazon
	"amazon/nova-2-lite-v1": true,
	// Arcee AI
	"arcee-ai/trinity-mini": true,
	// Cognitive Computations
	"dolphin-mistral-24b-venice-edition": true,
	// Google
	"google/gemma-3n-e2b-it":    true,
	"google/gemma-3n-e4b-it":    true,
	"google/gemma-3-4b-it":      true,
	"google/gemma-3-12b-it":     true,
	"google/gemma-3-27b-it":     true,
	"google/gemini-2.0-flash-exp": true,
	// KwaiPilot
	"kwaipilot/kat-coder-pro": true,
	// Meituan
	"meituan/longcat-flash-chat": true,
	// Meta Llama
	"meta-llama/llama-3.3-70b-instruct": true,
	"meta-llama/llama-3.2-3b-instruct":  true,
	// Mistral AI
	"mistralai/mistral-7b-instruct":           true,
	"mistralai/mistral-small-3.1-24b-instruct": true,
	// Moonshot AI
	"moonshotai/kimi-k2": true,
	// Nous Research
	"nousresearch/hermes-3-llama-3.1-405b": true,
	// NVIDIA
	"nvidia/nemotron-nano-12b-v2-vl": true,
	"nvidia/nemotron-nano-9b-v2":     true,
	// OpenAI
	"openai/gpt-oss-120b": true,
	"openai/gpt-oss-20b":  true,
	// Qwen
	"qwen/qwen-2.5-7b-instruct": true,
	"qwen/qwen3-coder":          true,
	"qwen/qwen3-4b":             true,
	"qwen/qwen3-235b-a22b":      true,
	// TNG Tech
	"tngtech/tng-r1t-chimera":      true,
	"tngtech/deepseek-r1t2-chimera": true,
	"tngtech/deepseek-r1t-chimera":  true,
	// Z-AI
	"glm-4.5-air": true,
}

// IsOpenRouterModel 检查是否为 OpenRouter 免费模型
func IsOpenRouterModel(model string) bool {
	return openRouterFreeModels[model]
}

// GetOpenRouterFreeModels 获取所有免费模型列表
func GetOpenRouterFreeModels() []string {
	models := make([]string, 0, len(openRouterFreeModels))
	for model := range openRouterFreeModels {
		models = append(models, model)
	}
	return models
}

// ChatCompletion 调用 OpenRouter API
func (s *OpenRouterService) ChatCompletion(ctx context.Context, request *models.ChatCompletionRequest) (<-chan interface{}, error) {
	// 构建请求体
	reqBody := map[string]interface{}{
		"model":    request.Model,
		"messages": s.convertMessages(request.Messages),
		"stream":   request.Stream,
	}
	
	if request.MaxTokens != nil {
		reqBody["max_tokens"] = *request.MaxTokens
	}
	if request.Temperature != nil {
		reqBody["temperature"] = *request.Temperature
	}
	if request.TopP != nil {
		reqBody["top_p"] = *request.TopP
	}
	if len(request.Stop) > 0 {
		reqBody["stop"] = request.Stop
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"model":  request.Model,
		"stream": request.Stream,
	}).Info("Sending OpenRouter API request")

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://cursor2api.com")
	httpReq.Header.Set("X-Title", "Cursor2API")

	// 发送请求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("OpenRouter API error")
		return nil, fmt.Errorf("OpenRouter API error: %d - %s", resp.StatusCode, string(body))
	}

	// 创建响应通道
	respChan := make(chan interface{}, 100)

	go func() {
		defer resp.Body.Close()
		defer close(respChan)

		if request.Stream {
			s.handleStreamResponse(resp.Body, respChan)
		} else {
			s.handleNonStreamResponse(resp.Body, respChan)
		}
	}()

	return respChan, nil
}


// convertMessages 转换消息格式
func (s *OpenRouterService) convertMessages(messages []models.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		converted := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.GetStringContent(),
		}
		result = append(result, converted)
	}
	return result
}

// handleStreamResponse 处理流式响应
func (s *OpenRouterService) handleStreamResponse(body io.Reader, respChan chan<- interface{}) {
	scanner := bufio.NewScanner(body)
	// 增加缓冲区大小以处理大响应
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		
		// 跳过空行
		if line == "" {
			continue
		}
		
		// 处理 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			
			// 检查是否结束
			if data == "[DONE]" {
				break
			}
			
			// 解析 JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				logrus.WithError(err).Debug("Failed to parse OpenRouter stream chunk")
				continue
			}
			
			// 提取内容
			if content := s.extractDeltaContent(chunk); content != "" {
				respChan <- content
			}
			
			// 检查是否结束
			if s.isFinished(chunk) {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Error("Error reading OpenRouter stream")
	}

	// 发送使用统计
	respChan <- models.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}
}

// handleNonStreamResponse 处理非流式响应
func (s *OpenRouterService) handleNonStreamResponse(body io.Reader, respChan chan<- interface{}) {
	var response map[string]interface{}
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		logrus.WithError(err).Error("Failed to decode OpenRouter response")
		return
	}

	// 提取内容
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					respChan <- content
				}
			}
		}
	}

	// 提取使用统计
	usage := models.Usage{}
	if usageData, ok := response["usage"].(map[string]interface{}); ok {
		if pt, ok := usageData["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := usageData["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		if tt, ok := usageData["total_tokens"].(float64); ok {
			usage.TotalTokens = int(tt)
		}
	}
	respChan <- usage
}

// extractDeltaContent 从流式响应中提取增量内容
func (s *OpenRouterService) extractDeltaContent(chunk map[string]interface{}) string {
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return ""
	}

	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return ""
	}

	content, _ := delta["content"].(string)
	return content
}

// isFinished 检查流式响应是否结束
func (s *OpenRouterService) isFinished(chunk map[string]interface{}) bool {
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return false
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return false
	}

	finishReason, ok := choice["finish_reason"].(string)
	return ok && finishReason != "" && finishReason != "null"
}


// GetOpenRouterFreeModelInfos 返回所有 OpenRouter 免费模型的详细信息
func GetOpenRouterFreeModelInfos() []models.ModelInfo {
	freeModels := []models.ModelInfo{
		// Alibaba
		{ID: "alibaba/tongyi-deepresearch-30b-a3b", Name: "🆓 Alibaba Tongyi DeepResearch 30B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// AllenAI
		{ID: "allenai/olmo-3-32b-think", Name: "🆓 AllenAI OLMo 3 32B Think", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Amazon
		{ID: "amazon/nova-2-lite-v1", Name: "🆓 Amazon Nova 2 Lite", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Arcee AI
		{ID: "arcee-ai/trinity-mini", Name: "🆓 Arcee AI Trinity Mini", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Cognitive Computations
		{ID: "dolphin-mistral-24b-venice-edition", Name: "🆓 Dolphin Mistral 24B Venice", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Google
		{ID: "google/gemma-3n-e2b-it", Name: "🆓 Google Gemma 3N E2B IT", Provider: "openrouter-free", ContextWindow: 8192, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "google/gemma-3n-e4b-it", Name: "🆓 Google Gemma 3N E4B IT", Provider: "openrouter-free", ContextWindow: 8192, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "google/gemma-3-4b-it", Name: "🆓 Google Gemma 3 4B IT", Provider: "openrouter-free", ContextWindow: 8192, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "google/gemma-3-12b-it", Name: "🆓 Google Gemma 3 12B IT", Provider: "openrouter-free", ContextWindow: 8192, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "google/gemma-3-27b-it", Name: "🆓 Google Gemma 3 27B IT", Provider: "openrouter-free", ContextWindow: 8192, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "google/gemini-2.0-flash-exp", Name: "🆓 Google Gemini 2.0 Flash Exp", Provider: "openrouter-free", ContextWindow: 1048576, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// KwaiPilot
		{ID: "kwaipilot/kat-coder-pro", Name: "🆓 KwaiPilot Kat Coder Pro", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Meituan
		{ID: "meituan/longcat-flash-chat", Name: "🆓 Meituan LongCat Flash Chat", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Meta Llama
		{ID: "meta-llama/llama-3.3-70b-instruct", Name: "🆓 Meta Llama 3.3 70B Instruct", Provider: "openrouter-free", ContextWindow: 131072, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "meta-llama/llama-3.2-3b-instruct", Name: "🆓 Meta Llama 3.2 3B Instruct", Provider: "openrouter-free", ContextWindow: 131072, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Mistral AI
		{ID: "mistralai/mistral-7b-instruct", Name: "🆓 Mistral 7B Instruct", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "mistralai/mistral-small-3.1-24b-instruct", Name: "🆓 Mistral Small 3.1 24B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Moonshot AI
		{ID: "moonshotai/kimi-k2", Name: "🆓 Moonshot Kimi K2", Provider: "openrouter-free", ContextWindow: 131072, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Nous Research
		{ID: "nousresearch/hermes-3-llama-3.1-405b", Name: "🆓 Nous Hermes 3 Llama 3.1 405B", Provider: "openrouter-free", ContextWindow: 131072, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// NVIDIA
		{ID: "nvidia/nemotron-nano-12b-v2-vl", Name: "🆓 NVIDIA Nemotron Nano 12B V2 VL", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "nvidia/nemotron-nano-9b-v2", Name: "🆓 NVIDIA Nemotron Nano 9B V2", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// OpenAI
		{ID: "openai/gpt-oss-120b", Name: "🆓 OpenAI GPT OSS 120B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "openai/gpt-oss-20b", Name: "🆓 OpenAI GPT OSS 20B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Qwen
		{ID: "qwen/qwen-2.5-7b-instruct", Name: "🆓 Qwen 2.5 7B Instruct", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "qwen/qwen3-coder", Name: "🆓 Qwen 3 Coder", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "qwen/qwen3-4b", Name: "🆓 Qwen 3 4B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "qwen/qwen3-235b-a22b", Name: "🆓 Qwen 3 235B A22B", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// TNG Tech
		{ID: "tngtech/tng-r1t-chimera", Name: "🆓 TNG R1T Chimera", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "tngtech/deepseek-r1t2-chimera", Name: "🆓 TNG DeepSeek R1T2 Chimera", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		{ID: "tngtech/deepseek-r1t-chimera", Name: "🆓 TNG DeepSeek R1T Chimera", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
		// Z-AI
		{ID: "glm-4.5-air", Name: "🆓 GLM 4.5 Air", Provider: "openrouter-free", ContextWindow: 32768, InputPrice: 0, OutputPrice: 0, IsAvailable: true},
	}
	return freeModels
}
