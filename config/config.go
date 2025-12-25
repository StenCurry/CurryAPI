package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config 应用程序配置结构
type Config struct {
	// 服务器配置
	Port  int  `json:"port"`
	Debug bool `json:"debug"`

	// API配置
	APIKey             string `json:"api_key"`
	Models             string `json:"models"`
	SystemPromptInject string `json:"system_prompt_inject"`
	Timeout            int    `json:"timeout"`
	MaxInputLength     int    `json:"max_input_length"`

	// 限流配置
	RateLimitRPS   int `json:"rate_limit_rps"`
	RateLimitBurst int `json:"rate_limit_burst"`

	// SMTP邮件配置
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from"`

	// 数据库配置
	DBType            string `json:"db_type"`             // sqlite 或 mysql
	DatabasePath      string `json:"database_path"`       // SQLite 数据库文件路径
	MySQLHost         string `json:"mysql_host"`          // MySQL 主机地址
	MySQLPort         int    `json:"mysql_port"`          // MySQL 端口
	MySQLUser         string `json:"mysql_user"`          // MySQL 用户名
	MySQLPassword     string `json:"mysql_password"`      // MySQL 密码
	MySQLDatabase     string `json:"mysql_database"`      // MySQL 数据库名
	DBMaxOpenConns    int    `json:"db_max_open_conns"`   // 最大打开连接数
	DBMaxIdleConns    int    `json:"db_max_idle_conns"`   // 最大空闲连接数
	DBConnMaxLifetime string `json:"db_conn_max_lifetime"` // 连接最大生命周期
	DBConnMaxIdleTime string `json:"db_conn_max_idle_time"` // 空闲连接最大生命周期

	// Cursor相关配置
	ScriptURL string `json:"script_url"`
	FP        FP     `json:"fp"`
	
	// Quota management configuration
	Quota QuotaConfig `json:"quota"`
	
	// Usage tracking configuration
	UsageTracking UsageTrackingConfig `json:"usage_tracking"`
	
	// AI Provider configurations
	Providers ProviderConfig `json:"providers"`
}

// FP 指纹配置结构
type FP struct {
	UserAgent               string `json:"userAgent"`
	UNMASKED_VENDOR_WEBGL   string `json:"unmaskedVendorWebgl"`
	UNMASKED_RENDERER_WEBGL string `json:"unmaskedRendererWebgl"`
}

// QuotaConfig 配额管理配置结构
type QuotaConfig struct {
	Enabled              bool    `json:"enabled"`                // Enable/disable quota management
	DefaultFreeQuota     int64   `json:"default_free_quota"`     // Default for free accounts
	DefaultProQuota      int64   `json:"default_pro_quota"`      // Default for pro accounts
	LowQuotaThreshold    float64 `json:"low_quota_threshold"`    // Percentage threshold for "low" status
	ResetHourUTC         int     `json:"reset_hour_utc"`         // Hour for daily reset (0 = midnight)
	EstimationMultiplier float64 `json:"estimation_multiplier"`  // Multiplier for token estimation
	MaxRetries           int     `json:"max_retries"`            // Max retries for DB writes
	RetryBackoffMs       int     `json:"retry_backoff_ms"`       // Initial backoff for retries (ms)
}

// UsageTrackingConfig 使用跟踪配置结构
type UsageTrackingConfig struct {
	Enabled        bool `json:"enabled"`          // Enable/disable usage tracking
	ChannelSize    int  `json:"channel_size"`     // Size of the buffered channel
	BatchSize      int  `json:"batch_size"`       // Number of records to batch before writing
	FlushInterval  int  `json:"flush_interval"`   // How often to flush batches (seconds)
	MaxRetries     int  `json:"max_retries"`      // Maximum number of retry attempts
	RetryBackoffMs int  `json:"retry_backoff_ms"` // Initial backoff for retries (ms)
	RetentionDays  int  `json:"retention_days"`   // Number of days to retain usage records
	CleanupHour    int  `json:"cleanup_hour"`     // Hour of day to run cleanup (0-23, UTC)
	CleanupMinute  int  `json:"cleanup_minute"`   // Minute of hour to run cleanup (0-59)
}

// OpenAIConfig OpenAI provider configuration
type OpenAIConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// AnthropicConfig Anthropic provider configuration
type AnthropicConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// GoogleConfig Google AI provider configuration
type GoogleConfig struct {
	APIKey string `json:"api_key"`
}

// DeepSeekConfig DeepSeek provider configuration
type DeepSeekConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// ProviderConfig AI provider configurations
type ProviderConfig struct {
	OpenAI    OpenAIConfig    `json:"openai"`
	Anthropic AnthropicConfig `json:"anthropic"`
	Google    GoogleConfig    `json:"google"`
	DeepSeek  DeepSeekConfig  `json:"deepseek"`
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 尝试加载.env文件
	if err := godotenv.Load(); err != nil {
		logrus.Debug("No .env file found, using environment variables")
	}

	config := &Config{
		// 设置默认值
		Port:               getEnvAsInt("PORT", 8002),
		Debug:              getEnvAsBool("DEBUG", false),
		APIKey:             getEnv("API_KEY", "0000"),
		Models:             getEnv("MODELS", "gpt-5.2,gpt-5,gpt-5.1,gpt-4o,claude-3.5-sonnet"),
		SystemPromptInject: getEnv("SYSTEM_PROMPT_INJECT", ""),
		Timeout:            getEnvAsInt("TIMEOUT", 30),
		MaxInputLength:     getEnvAsInt("MAX_INPUT_LENGTH", 200000),
		RateLimitRPS:       getEnvAsInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:     getEnvAsInt("RATE_LIMIT_BURST", 20),
		// SMTP配置（163邮箱）
		SMTPHost:     getEnv("SMTP_HOST", "smtp.163.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 465),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		// 数据库配置
		DBType:            getEnv("DB_TYPE", "sqlite"), // 默认使用 SQLite
		DatabasePath:      getEnv("DATABASE_PATH", "data.db"),
		MySQLHost:         getEnv("MYSQL_HOST", "localhost"),
		MySQLPort:         getEnvAsInt("MYSQL_PORT", 3306),
		MySQLUser:         getEnv("MYSQL_USER", "root"),
		MySQLPassword:     getEnv("MYSQL_PASSWORD", ""),
		MySQLDatabase:     getEnv("MYSQL_DATABASE", "Curry2API"),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: getEnv("DB_CONN_MAX_LIFETIME", "5m"),
		DBConnMaxIdleTime: getEnv("DB_CONN_MAX_IDLE_TIME", "10m"),
		ScriptURL:    getEnv("SCRIPT_URL", "https://cursor.com/_next/static/chunks/pages/_app.js"),
		FP: FP{
			UserAgent:               getEnv("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"),
			UNMASKED_VENDOR_WEBGL:   getEnv("UNMASKED_VENDOR_WEBGL", "Google Inc. (Intel)"),
			UNMASKED_RENDERER_WEBGL: getEnv("UNMASKED_RENDERER_WEBGL", "ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11)"),
		},
		// Quota management configuration
		Quota: QuotaConfig{
			Enabled:              getEnvAsBool("QUOTA_ENABLED", true),
			DefaultFreeQuota:     getEnvAsInt64("QUOTA_DEFAULT_FREE", 100000),
			DefaultProQuota:      getEnvAsInt64("QUOTA_DEFAULT_PRO", 500000),
			LowQuotaThreshold:    getEnvAsFloat64("QUOTA_LOW_THRESHOLD", 0.8),
			ResetHourUTC:         getEnvAsInt("QUOTA_RESET_HOUR_UTC", 0),
			EstimationMultiplier: getEnvAsFloat64("QUOTA_ESTIMATION_MULTIPLIER", 1.5),
			MaxRetries:           getEnvAsInt("QUOTA_MAX_RETRIES", 3),
			RetryBackoffMs:       getEnvAsInt("QUOTA_RETRY_BACKOFF_MS", 100),
		},
		// Usage tracking configuration
		UsageTracking: UsageTrackingConfig{
			Enabled:        getEnvAsBool("USAGE_TRACKING_ENABLED", true),
			ChannelSize:    getEnvAsInt("USAGE_CHANNEL_SIZE", 1000),
			BatchSize:      getEnvAsInt("USAGE_BATCH_SIZE", 100),
			FlushInterval:  getEnvAsInt("USAGE_FLUSH_INTERVAL", 5),
			MaxRetries:     getEnvAsInt("USAGE_MAX_RETRIES", 3),
			RetryBackoffMs: getEnvAsInt("USAGE_RETRY_BACKOFF_MS", 100),
			RetentionDays:  getEnvAsInt("USAGE_RETENTION_DAYS", 90),
			CleanupHour:    getEnvAsInt("USAGE_CLEANUP_HOUR", 3),
			CleanupMinute:  getEnvAsInt("USAGE_CLEANUP_MINUTE", 0),
		},
		// AI Provider configurations
		Providers: ProviderConfig{
			OpenAI: OpenAIConfig{
				APIKey:  getEnv("OPENAI_API_KEY", ""),
				BaseURL: getEnv("OPENAI_API_BASE", "https://api.openai.com/v1"),
			},
			Anthropic: AnthropicConfig{
				APIKey:  getEnv("ANTHROPIC_API_KEY", ""),
				BaseURL: getEnv("ANTHROPIC_API_BASE", "https://api.anthropic.com/v1"),
			},
			Google: GoogleConfig{
				APIKey: getEnv("GOOGLE_AI_API_KEY", ""),
			},
			DeepSeek: DeepSeekConfig{
				APIKey:  getEnv("DEEPSEEK_API_KEY", ""),
				BaseURL: getEnv("DEEPSEEK_API_BASE", "https://api.deepseek.com/v1"),
			},
		},
	}

	// 验证必要的配置
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate 验证配置
func (c *Config) validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.APIKey == "" {
		return fmt.Errorf("API_KEY is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxInputLength <= 0 {
		return fmt.Errorf("max input length must be positive")
	}

	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("rate limit RPS must be positive")
	}

	if c.RateLimitBurst <= 0 {
		return fmt.Errorf("rate limit burst must be positive")
	}

	return nil
}

// GetModels 获取模型列表
func (c *Config) GetModels() []string {
	models := strings.Split(c.Models, ",")
	result := make([]string, 0, len(models))
	for _, model := range models {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetAvailableProviders returns list of providers with valid API keys
func (c *Config) GetAvailableProviders() []string {
	providers := make([]string, 0, 4)
	
	if c.Providers.OpenAI.APIKey != "" {
		providers = append(providers, "openai")
	}
	if c.Providers.Anthropic.APIKey != "" {
		providers = append(providers, "anthropic")
	}
	if c.Providers.Google.APIKey != "" {
		providers = append(providers, "google")
	}
	if c.Providers.DeepSeek.APIKey != "" {
		providers = append(providers, "deepseek")
	}
	
	// Cursor is always available as it uses the existing system
	providers = append(providers, "cursor")
	
	return providers
}

// NormalizeModelName 标准化模型名称，将完整的模型标识符映射到配置中的简短名称
func (c *Config) NormalizeModelName(model string) string {
	// 模型名称映射表：完整标识符 -> 配置中的简短名称
	modelMappings := map[string]string{
		// Claude 3.5 Sonnet (旧版本)
		"claude-3-5-sonnet-20241022":  "claude-3.5-sonnet",
		"claude-3-5-sonnet-20240620":  "claude-3.5-sonnet",
		
		// Claude 3.5 Haiku (旧版本)
		"claude-3-5-haiku-20241022":   "claude-3.5-haiku",
		
		// Claude 3 Opus
		"claude-3-opus-20240229":      "claude-3.7-sonnet",
		
		// Claude 3 Sonnet
		"claude-3-sonnet-20240229":    "claude-3.7-sonnet",
		
		// Claude 3 Haiku
		"claude-3-haiku-20240307":     "claude-3.5-haiku",
		
		// Claude 4 Sonnet 系列
		"claude-4-sonnet":             "claude-4-sonnet",
		"claude-sonnet-4-20250514":    "claude-4-sonnet",
		
		// Claude 4.5 Sonnet 系列 (修正映射)
		"claude-4.5-sonnet":           "claude-4.5-sonnet",
		"claude-4-5-sonnet":           "claude-4.5-sonnet",
		"claude-sonnet-4-5-20250929":  "claude-4.5-sonnet",
		
		// Claude 4 Opus 系列
		"claude-4-opus":               "claude-4-opus",
		"claude-opus-4-20250514":      "claude-4-opus",
		
		// Claude 4.1 Opus 系列
		"claude-4.1-opus":             "claude-4.1-opus",
		"claude-4-1-opus":             "claude-4.1-opus",
		"claude-opus-4-1-20250620":    "claude-4.1-opus",
		
		// Claude 4.5 Opus 系列 (新增)
		"claude-4.5-opus":             "claude-4.5-opus",
		"claude-4-5-opus":             "claude-4.5-opus",
		"claude-opus-4-5-20251101":    "claude-4.5-opus",
		
		// Claude 4.5 Haiku 系列 (修正映射)
		"claude-4.5-haiku":            "claude-4.5-haiku",
		"claude-4-5-haiku":            "claude-4.5-haiku",
		"claude-haiku-4-5-20251001":   "claude-4.5-haiku",
		
		// GPT 系列（支持各种变体）
		"gpt-5.2":                     "gpt-5.2",
		"gpt-5-2":                     "gpt-5.2",
		"gpt-5.1":                     "gpt-5.1",
		"gpt-5.1-codex":               "gpt-5.1-codex",
		"gpt-5.1-codex-max":           "gpt-5.1-codex-max",
		"gpt-5-1-codex-max":           "gpt-5.1-codex-max",
		"gpt-5-codex":                 "gpt-5-codex",
		"gpt-5":                       "gpt-5",
		"gpt-5-mini":                  "gpt-5-mini",
		"gpt-5-nano":                  "gpt-5-nano",
		"gpt-4.1":                     "gpt-4.1",
		"gpt-4o":                      "gpt-4o",
		"gpt-4":                       "gpt-4o",
		"gpt-4-turbo":                 "gpt-4o",
		"gpt-3.5-turbo":               "gpt-5-mini",
		
		// O 系列
		"o3":                          "o3",
		"o4-mini":                     "o4-mini",
		"o1":                          "o3",
		"o1-mini":                     "o4-mini",
		
		// 其他模型
		"deepseek-r1":                 "deepseek-r1",
		"deepseek-v3.1":               "deepseek-v3.1",
		"gemini-2.5-pro":              "gemini-2.5-pro",
		"gemini-2.5-flash":            "gemini-2.5-flash",
		"gemini-3-pro-preview":        "gemini-3-pro-preview",
		"gemini-3-pro":                "gemini-3-pro-preview",
		
		// 其他模型
		"kimi-k2-instruct":            "kimi-k2-instruct",
		"grok-3":                      "grok-3",
		"grok-3-mini":                 "grok-3-mini",
		"grok-4":                      "grok-4",
		"code-supernova-1-million":    "code-supernova-1-million",
	}
	
	// 如果有映射，返回映射后的名称
	if normalized, exists := modelMappings[model]; exists {
		return normalized
	}
	
	// 否则返回原始名称
	return model
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
	"google/gemma-3n-e2b-it":      true,
	"google/gemma-3n-e4b-it":      true,
	"google/gemma-3-4b-it":        true,
	"google/gemma-3-12b-it":       true,
	"google/gemma-3-27b-it":       true,
	"google/gemini-2.0-flash-exp": true,
	// KwaiPilot
	"kwaipilot/kat-coder-pro": true,
	// Meituan
	"meituan/longcat-flash-chat": true,
	// Meta Llama
	"meta-llama/llama-3.3-70b-instruct": true,
	"meta-llama/llama-3.2-3b-instruct":  true,
	// Mistral AI
	"mistralai/mistral-7b-instruct":            true,
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
	"tngtech/tng-r1t-chimera":       true,
	"tngtech/deepseek-r1t2-chimera": true,
	"tngtech/deepseek-r1t-chimera":  true,
	// Z-AI
	"glm-4.5-air": true,
}

// IsOpenRouterFreeModel 检查是否为 OpenRouter 免费模型
func IsOpenRouterFreeModel(model string) bool {
	return openRouterFreeModels[model]
}

// GetOpenRouterFreeModels 获取所有 OpenRouter 免费模型列表
func GetOpenRouterFreeModels() []string {
	models := make([]string, 0, len(openRouterFreeModels))
	for model := range openRouterFreeModels {
		models = append(models, model)
	}
	return models
}

// IsValidModel 检查模型是否有效（支持完整标识符和简短名称）
func (c *Config) IsValidModel(model string) bool {
	// 检查是否为 OpenRouter 免费模型
	if IsOpenRouterFreeModel(model) {
		return true
	}
	
	// 先尝试标准化模型名称
	normalizedModel := c.NormalizeModelName(model)
	
	validModels := c.GetModels()
	for _, validModel := range validModels {
		if validModel == normalizedModel || validModel == model {
			return true
		}
	}
	return false
}

// ToJSON 将配置序列化为JSON（用于调试）
func (c *Config) ToJSON() string {
	// 创建一个副本，隐藏敏感信息
	safeCfg := *c
	safeCfg.APIKey = "***"

	data, err := json.MarshalIndent(safeCfg, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling config: %v", err)
	}
	return string(data)
}

// 辅助函数

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logrus.Warnf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// getEnvAsBool 获取环境变量并转换为bool
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		logrus.Warnf("Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}


// getEnvAsInt64 获取环境变量并转换为int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		logrus.Warnf("Invalid int64 value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// getEnvAsFloat64 获取环境变量并转换为float64
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		logrus.Warnf("Invalid float64 value for %s: %s, using default: %f", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
