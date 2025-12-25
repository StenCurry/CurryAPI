package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ModelMarketplaceInfo represents model information for the marketplace
// Requirements: 15.1-15.8
type ModelMarketplaceInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Provider      string   `json:"provider"`       // OpenAI, Anthropic, Google, etc.
	Tags          []string `json:"tags"`           // Fast, Powerful, Code, Vision
	BillingType   string   `json:"billing_type"`   // per_token, per_request
	EndpointType  string   `json:"endpoint_type"`  // chat, completion, embedding
	MaxTokens     int      `json:"max_tokens"`
	ContextWindow int      `json:"context_window"`
	Description   string   `json:"description"`
}

// GetModelMarketplace returns the full model marketplace data
func GetModelMarketplace() []ModelMarketplaceInfo {
	return []ModelMarketplaceInfo{
		// OpenAI GPT-5 Series
		{
			ID:            "gpt-5.2",
			Name:          "GPT-5.2",
			Provider:      "OpenAI",
			Tags:          []string{"Powerful", "Latest", "Premium"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 512000,
			Description:   "OpenAI's newest GPT-5.2 model with 512K context window and enhanced reasoning",
		},
		{
			ID:            "gpt-5",
			Name:          "GPT-5",
			Provider:      "OpenAI",
			Tags:          []string{"Powerful", "Latest"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 400000,
			Description:   "OpenAI's most advanced GPT-5 model with 400K context window",
		},
		{
			ID:            "gpt-5.1",
			Name:          "GPT-5.1",
			Provider:      "OpenAI",
			Tags:          []string{"Powerful", "Latest", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 500000,
			Description:   "GPT-5.1 with enhanced capabilities and 500K context window",
		},
		{
			ID:            "gpt-5-codex",
			Name:          "GPT-5 Codex",
			Provider:      "OpenAI",
			Tags:          []string{"Code", "Powerful"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 192000,
			Description:   "GPT-5 optimized for code generation and understanding",
		},
		{
			ID:            "gpt-5.1-codex",
			Name:          "GPT-5.1 Codex",
			Provider:      "OpenAI",
			Tags:          []string{"Code", "Powerful", "Latest"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 256000,
			Description:   "GPT-5.1 Codex with improved code generation",
		},
		{
			ID:            "gpt-5.1-codex-max",
			Name:          "GPT-5.1 Codex Max",
			Provider:      "OpenAI",
			Tags:          []string{"Code", "Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 256000,
			Description:   "Enhanced GPT-5 Codex with extended output and context",
		},
		{
			ID:            "gpt-5-mini",
			Name:          "GPT-5 Mini",
			Provider:      "OpenAI",
			Tags:          []string{"Fast", "Efficient"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 400000,
			Description:   "Smaller, faster version of GPT-5 for quick tasks",
		},
		{
			ID:            "gpt-5-nano",
			Name:          "GPT-5 Nano",
			Provider:      "OpenAI",
			Tags:          []string{"Fast", "Lightweight"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 400000,
			Description:   "Ultra-lightweight GPT-5 variant for simple tasks",
		},

		// OpenAI GPT-4 Series
		{
			ID:            "gpt-4.1",
			Name:          "GPT-4.1",
			Provider:      "OpenAI",
			Tags:          []string{"Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "GPT-4.1 with 1M context window for extensive documents",
		},
		{
			ID:            "gpt-4o",
			Name:          "GPT-4o",
			Provider:      "OpenAI",
			Tags:          []string{"Fast", "Multimodal", "Vision"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 128000,
			Description:   "GPT-4 Omni with multimodal capabilities",
		},

		// Anthropic Claude Series
		{
			ID:            "claude-3.5-sonnet",
			Name:          "Claude 3.5 Sonnet",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Balanced"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 200000,
			Description:   "Balanced Claude model for general tasks",
		},
		{
			ID:            "claude-3.5-haiku",
			Name:          "Claude 3.5 Haiku",
			Provider:      "Anthropic",
			Tags:          []string{"Fast", "Efficient"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 200000,
			Description:   "Fast and efficient Claude model",
		},
		{
			ID:            "claude-3.7-sonnet",
			Name:          "Claude 3.7 Sonnet",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Latest"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 200000,
			Description:   "Latest Claude 3.7 Sonnet with improved capabilities",
		},
		{
			ID:            "claude-4-sonnet",
			Name:          "Claude 4 Sonnet",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Claude 4 Sonnet with 1M context window",
		},
		{
			ID:            "claude-4.5-sonnet",
			Name:          "Claude 4.5 Sonnet",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Latest", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Latest Claude 4.5 Sonnet with enhanced capabilities",
		},
		{
			ID:            "claude-4-opus",
			Name:          "Claude 4 Opus",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Premium"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 200000,
			Description:   "Claude 4 Opus - premium model for complex tasks",
		},
		{
			ID:            "claude-4.1-opus",
			Name:          "Claude 4.1 Opus",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Premium"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 200000,
			Description:   "Claude 4.1 Opus with improved reasoning",
		},
		{
			ID:            "claude-4.5-opus",
			Name:          "Claude 4.5 Opus",
			Provider:      "Anthropic",
			Tags:          []string{"Powerful", "Premium", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Latest Claude 4.5 Opus with 1M context",
		},
		{
			ID:            "claude-4.5-haiku",
			Name:          "Claude 4.5 Haiku",
			Provider:      "Anthropic",
			Tags:          []string{"Fast", "Efficient", "Latest"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 200000,
			Description:   "Fast Claude 4.5 Haiku for quick responses",
		},
		{
			ID:            "claude-code-1m",
			Name:          "Claude Code 1M",
			Provider:      "Anthropic",
			Tags:          []string{"Code", "Extended", "Powerful"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Claude Code with 1M context window for large codebases",
		},

		// Google Gemini Series
		{
			ID:            "gemini-2.5-pro",
			Name:          "Gemini 2.5 Pro",
			Provider:      "Google",
			Tags:          []string{"Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Google Gemini 2.5 Pro with 1M context window",
		},
		{
			ID:            "gemini-2.5-flash",
			Name:          "Gemini 2.5 Flash",
			Provider:      "Google",
			Tags:          []string{"Fast", "Efficient"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Fast Gemini model optimized for speed",
		},
		{
			ID:            "gemini-3-pro-preview",
			Name:          "Gemini 3 Pro Preview",
			Provider:      "Google",
			Tags:          []string{"Powerful", "Preview", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 2000000,
			Description:   "Preview of Gemini 3 Pro with 2M context window",
		},

		// OpenAI O-Series (Reasoning)
		{
			ID:            "o3",
			Name:          "O3",
			Provider:      "OpenAI",
			Tags:          []string{"Reasoning", "Powerful"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 200000,
			Description:   "OpenAI O3 reasoning model for complex problems",
		},
		{
			ID:            "o4-mini",
			Name:          "O4 Mini",
			Provider:      "OpenAI",
			Tags:          []string{"Reasoning", "Fast"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 200000,
			Description:   "Compact O4 reasoning model for quick analysis",
		},

		// DeepSeek Series
		{
			ID:            "deepseek-r1",
			Name:          "DeepSeek R1",
			Provider:      "DeepSeek",
			Tags:          []string{"Reasoning", "Code"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 128000,
			Description:   "DeepSeek R1 reasoning model",
		},
		{
			ID:            "deepseek-v3.1",
			Name:          "DeepSeek V3.1",
			Provider:      "DeepSeek",
			Tags:          []string{"Powerful", "Code"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 128000,
			Description:   "DeepSeek V3.1 general purpose model",
		},

		// Moonshot AI
		{
			ID:            "kimi-k2-instruct",
			Name:          "Kimi K2 Instruct",
			Provider:      "Moonshot",
			Tags:          []string{"Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 256000,
			Description:   "Moonshot Kimi K2 with instruction following",
		},

		// xAI Grok Series
		{
			ID:            "grok-3",
			Name:          "Grok 3",
			Provider:      "xAI",
			Tags:          []string{"Powerful", "Extended"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "xAI Grok 3 with 1M context window",
		},
		{
			ID:            "grok-3-mini",
			Name:          "Grok 3 Mini",
			Provider:      "xAI",
			Tags:          []string{"Fast", "Efficient"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 131072,
			Description:   "Compact Grok 3 for quick tasks",
		},
		{
			ID:            "grok-4",
			Name:          "Grok 4",
			Provider:      "xAI",
			Tags:          []string{"Powerful", "Latest"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 256000,
			Description:   "Latest xAI Grok 4 model",
		},

		// Code Supernova
		{
			ID:            "code-supernova-1-million",
			Name:          "Code Supernova 1M",
			Provider:      "Code Supernova",
			Tags:          []string{"Code", "Extended", "Powerful"},
			BillingType:   "per_token",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1000000,
			Description:   "Code Supernova with 1M context for large codebases",
		},

		// ========== OpenRouter 免费模型 ==========
		// Alibaba
		{
			ID:            "alibaba/tongyi-deepresearch-30b-a3b",
			Name:          "🆓 Alibaba Tongyi DeepResearch 30B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Research"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Alibaba Tongyi DeepResearch 30B - 免费模型",
		},
		// AllenAI
		{
			ID:            "allenai/olmo-3-32b-think",
			Name:          "🆓 AllenAI OLMo 3 32B Think",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Reasoning"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "AllenAI OLMo 3 32B Think - 免费推理模型",
		},
		// Amazon
		{
			ID:            "amazon/nova-2-lite-v1",
			Name:          "🆓 Amazon Nova 2 Lite",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Amazon Nova 2 Lite - 免费轻量模型",
		},
		// Arcee AI
		{
			ID:            "arcee-ai/trinity-mini",
			Name:          "🆓 Arcee AI Trinity Mini",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Arcee AI Trinity Mini - 免费迷你模型",
		},
		// Cognitive Computations
		{
			ID:            "dolphin-mistral-24b-venice-edition",
			Name:          "🆓 Dolphin Mistral 24B Venice",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Uncensored"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Dolphin Mistral 24B Venice Edition - 免费无审查模型",
		},
		// Google Gemma
		{
			ID:            "google/gemma-3n-e2b-it",
			Name:          "🆓 Google Gemma 3N E2B IT",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 8192,
			Description:   "Google Gemma 3N E2B IT - 免费轻量模型",
		},
		{
			ID:            "google/gemma-3n-e4b-it",
			Name:          "🆓 Google Gemma 3N E4B IT",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 8192,
			Description:   "Google Gemma 3N E4B IT - 免费轻量模型",
		},
		{
			ID:            "google/gemma-3-4b-it",
			Name:          "🆓 Google Gemma 3 4B IT",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 8192,
			Description:   "Google Gemma 3 4B IT - 免费4B模型",
		},
		{
			ID:            "google/gemma-3-12b-it",
			Name:          "🆓 Google Gemma 3 12B IT",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Balanced"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 8192,
			Description:   "Google Gemma 3 12B IT - 免费12B模型",
		},
		{
			ID:            "google/gemma-3-27b-it",
			Name:          "🆓 Google Gemma 3 27B IT",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Powerful"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 8192,
			Description:   "Google Gemma 3 27B IT - 免费27B模型",
		},
		{
			ID:            "google/gemini-2.0-flash-exp",
			Name:          "🆓 Google Gemini 2.0 Flash Exp",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast", "Extended"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     8192,
			ContextWindow: 1048576,
			Description:   "Google Gemini 2.0 Flash Experimental - 免费1M上下文",
		},
		// KwaiPilot
		{
			ID:            "kwaipilot/kat-coder-pro",
			Name:          "🆓 KwaiPilot Kat Coder Pro",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Code"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "KwaiPilot Kat Coder Pro - 免费代码模型",
		},
		// Meituan
		{
			ID:            "meituan/longcat-flash-chat",
			Name:          "🆓 Meituan LongCat Flash Chat",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Meituan LongCat Flash Chat - 免费快速模型",
		},
		// Meta Llama
		{
			ID:            "meta-llama/llama-3.3-70b-instruct",
			Name:          "🆓 Meta Llama 3.3 70B Instruct",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Powerful"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 131072,
			Description:   "Meta Llama 3.3 70B Instruct - 免费70B大模型",
		},
		{
			ID:            "meta-llama/llama-3.2-3b-instruct",
			Name:          "🆓 Meta Llama 3.2 3B Instruct",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 131072,
			Description:   "Meta Llama 3.2 3B Instruct - 免费轻量模型",
		},
		// Mistral AI
		{
			ID:            "mistralai/mistral-7b-instruct",
			Name:          "🆓 Mistral 7B Instruct",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Mistral 7B Instruct - 免费7B模型",
		},
		{
			ID:            "mistralai/mistral-small-3.1-24b-instruct",
			Name:          "🆓 Mistral Small 3.1 24B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Balanced"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Mistral Small 3.1 24B Instruct - 免费24B模型",
		},
		// Moonshot AI
		{
			ID:            "moonshotai/kimi-k2",
			Name:          "🆓 Moonshot Kimi K2",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Extended"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 131072,
			Description:   "Moonshot Kimi K2 - 免费长上下文模型",
		},
		// Nous Research
		{
			ID:            "nousresearch/hermes-3-llama-3.1-405b",
			Name:          "🆓 Nous Hermes 3 Llama 3.1 405B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Powerful"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 131072,
			Description:   "Nous Hermes 3 Llama 3.1 405B - 免费405B超大模型",
		},
		// NVIDIA
		{
			ID:            "nvidia/nemotron-nano-12b-v2-vl",
			Name:          "🆓 NVIDIA Nemotron Nano 12B V2 VL",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Vision"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "NVIDIA Nemotron Nano 12B V2 VL - 免费视觉模型",
		},
		{
			ID:            "nvidia/nemotron-nano-9b-v2",
			Name:          "🆓 NVIDIA Nemotron Nano 9B V2",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "NVIDIA Nemotron Nano 9B V2 - 免费9B模型",
		},
		// OpenAI OSS
		{
			ID:            "openai/gpt-oss-120b",
			Name:          "🆓 OpenAI GPT OSS 120B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Powerful"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "OpenAI GPT OSS 120B - 免费开源120B模型",
		},
		{
			ID:            "openai/gpt-oss-20b",
			Name:          "🆓 OpenAI GPT OSS 20B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Balanced"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "OpenAI GPT OSS 20B - 免费开源20B模型",
		},
		// Qwen
		{
			ID:            "qwen/qwen-2.5-7b-instruct",
			Name:          "🆓 Qwen 2.5 7B Instruct",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Qwen 2.5 7B Instruct - 免费7B模型",
		},
		{
			ID:            "qwen/qwen3-coder",
			Name:          "🆓 Qwen 3 Coder",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Code"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Qwen 3 Coder - 免费代码模型",
		},
		{
			ID:            "qwen/qwen3-4b",
			Name:          "🆓 Qwen 3 4B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Fast"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Qwen 3 4B - 免费轻量模型",
		},
		{
			ID:            "qwen/qwen3-235b-a22b",
			Name:          "🆓 Qwen 3 235B A22B",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Powerful"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "Qwen 3 235B A22B - 免费超大模型",
		},
		// TNG Tech
		{
			ID:            "tngtech/tng-r1t-chimera",
			Name:          "🆓 TNG R1T Chimera",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Reasoning"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "TNG R1T Chimera - 免费推理模型",
		},
		{
			ID:            "tngtech/deepseek-r1t2-chimera",
			Name:          "🆓 TNG DeepSeek R1T2 Chimera",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Reasoning"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "TNG DeepSeek R1T2 Chimera - 免费推理模型",
		},
		{
			ID:            "tngtech/deepseek-r1t-chimera",
			Name:          "🆓 TNG DeepSeek R1T Chimera",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Reasoning"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "TNG DeepSeek R1T Chimera - 免费推理模型",
		},
		// Z-AI
		{
			ID:            "glm-4.5-air",
			Name:          "🆓 GLM 4.5 Air",
			Provider:      "OpenRouter Free",
			Tags:          []string{"Free", "Balanced"},
			BillingType:   "free",
			EndpointType:  "chat",
			MaxTokens:     4096,
			ContextWindow: 32768,
			Description:   "GLM 4.5 Air - 免费智谱模型",
		},
	}
}


// GetModelMarketplaceHandler returns all available models for the marketplace
// GET /api/models/marketplace
// Query params: provider (filter by provider), tag (filter by tag), endpoint_type (filter by endpoint type)
// Requirements: 15.1-15.8
func GetModelMarketplaceHandler(c *gin.Context) {
	models := GetModelMarketplace()

	// Get filter parameters
	providerFilter := c.Query("provider")
	tagFilter := c.Query("tag")
	endpointTypeFilter := c.Query("endpoint_type")

	// Apply filters if provided
	if providerFilter != "" || tagFilter != "" || endpointTypeFilter != "" {
		filteredModels := make([]ModelMarketplaceInfo, 0)
		for _, model := range models {
			// Filter by provider (case-insensitive)
			if providerFilter != "" && !strings.EqualFold(model.Provider, providerFilter) {
				continue
			}

			// Filter by endpoint type (case-insensitive)
			if endpointTypeFilter != "" && !strings.EqualFold(model.EndpointType, endpointTypeFilter) {
				continue
			}

			// Filter by tag (case-insensitive, check if any tag matches)
			if tagFilter != "" {
				hasTag := false
				for _, tag := range model.Tags {
					if strings.EqualFold(tag, tagFilter) {
						hasTag = true
						break
					}
				}
				if !hasTag {
					continue
				}
			}

			filteredModels = append(filteredModels, model)
		}
		models = filteredModels
	}

	// Get unique providers for filter options
	providerSet := make(map[string]bool)
	tagSet := make(map[string]bool)
	endpointTypeSet := make(map[string]bool)

	allModels := GetModelMarketplace()
	for _, model := range allModels {
		providerSet[model.Provider] = true
		endpointTypeSet[model.EndpointType] = true
		for _, tag := range model.Tags {
			tagSet[tag] = true
		}
	}

	providers := make([]string, 0, len(providerSet))
	for provider := range providerSet {
		providers = append(providers, provider)
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	endpointTypes := make([]string, 0, len(endpointTypeSet))
	for endpointType := range endpointTypeSet {
		endpointTypes = append(endpointTypes, endpointType)
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
		"filters": gin.H{
			"providers":      providers,
			"tags":           tags,
			"endpoint_types": endpointTypes,
		},
	})
}

// GetModelDetailHandler returns detailed information for a specific model
// GET /api/models/marketplace/:id
// Requirements: 15.7
func GetModelDetailHandler(c *gin.Context) {
	modelID := c.Param("id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model ID is required",
		})
		return
	}

	models := GetModelMarketplace()
	for _, model := range models {
		if model.ID == modelID {
			c.JSON(http.StatusOK, model)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Model not found",
	})
}
