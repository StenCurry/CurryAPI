package main

import (
	"context"
	"Curry2API-go/config"
	"Curry2API-go/database"
	"Curry2API-go/handlers"
	"Curry2API-go/middleware"
	"Curry2API-go/services"
	"Curry2API-go/utils"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Init(cfg); err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	db, err := database.GetDB()
	if err != nil {
		logrus.Fatalf("Failed to get database: %v", err)
	}
	defer db.Close()

	// 环境变量迁移（仅首次）
	if err := database.MigrateFromEnv(); err != nil {
		logrus.Warnf("Failed to migrate from env: %v", err)
	}

	// 设置日志级别
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
	
	// 添加缓存控制中间件（防止API响应被缓存）
	router.Use(func(c *gin.Context) {
		// 对所有API请求添加no-cache头
		path := c.Request.URL.Path
		isAPIPath := false
		
		if len(path) >= 3 && path[:3] == "/v1" {
			isAPIPath = true
		} else if len(path) >= 4 && path[:4] == "/api" {
			isAPIPath = true
		} else if len(path) >= 5 && path[:5] == "/auth" {
			isAPIPath = true
		} else if len(path) >= 6 && path[:6] == "/admin" {
			isAPIPath = true
		} else if len(path) >= 8 && path[:8] == "/profile" {
			isAPIPath = true
		} else if len(path) >= 14 && path[:14] == "/announcements" {
			isAPIPath = true
		}
		
		if isAPIPath {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// 初始化邮件服务
	handlers.InitEmailService(cfg)

	// 初始化 Turnstile 服务（必需）
	turnstileSecretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	if turnstileSecretKey == "" {
		logrus.Fatal("TURNSTILE_SECRET_KEY is required but not configured. Please set it in .env file")
	}
	handlers.InitTurnstileService(turnstileSecretKey)
	logrus.Info("Turnstile service initialized successfully")

	// 初始化 OAuth 加密
	if err := database.InitOAuthCrypto(); err != nil {
		logrus.Fatalf("Failed to initialize OAuth crypto: %v", err)
	}

	// 初始化数据加密（用于加密敏感数据如 cursor tokens）
	if err := utils.InitDataCrypto(); err != nil {
		logrus.Fatalf("Failed to initialize data crypto: %v", err)
	}

	// 初始化 OAuth 服务
	oauthConfig, err := services.LoadOAuthConfig()
	if err != nil {
		logrus.Warnf("Failed to load OAuth config: %v", err)
	}

	// Log usage tracking feature flag status
	if cfg.UsageTracking.Enabled {
		logrus.Info("Usage tracking is ENABLED")
	} else {
		logrus.Info("Usage tracking is DISABLED")
	}

	// Initialize usage tracker with config
	usageTrackerConfig := &services.UsageTrackerConfig{
		Enabled:        cfg.UsageTracking.Enabled,
		ChannelSize:    cfg.UsageTracking.ChannelSize,
		BatchSize:      cfg.UsageTracking.BatchSize,
		FlushInterval:  time.Duration(cfg.UsageTracking.FlushInterval) * time.Second,
		MaxRetries:     cfg.UsageTracking.MaxRetries,
		RetryBackoffMs: cfg.UsageTracking.RetryBackoffMs,
	}
	services.InitUsageTracker(usageTrackerConfig)

	// Initialize usage data cleanup service with config
	cleanupConfig := &services.CleanupConfig{
		Enabled:        cfg.UsageTracking.Enabled, // Cleanup follows tracking enabled state
		RetentionDays:  cfg.UsageTracking.RetentionDays,
		BatchSize:      1000,
		ScheduleHour:   cfg.UsageTracking.CleanupHour,
		ScheduleMinute: cfg.UsageTracking.CleanupMinute,
	}
	cleanupService := services.InitUsageCleanupService(cleanupConfig)
	cleanupService.Start()
	var oauthService *services.OAuthService
	var oauthHandler *handlers.OAuthHandler
	if oauthConfig != nil {
		// 设置数据库函数
		services.SetDatabaseFunctions(
			database.CreateOAuthState,
			database.VerifyOAuthState,
			database.DeleteOAuthState,
			database.CleanupExpiredOAuthStates,
		)
		
		oauthService = services.NewOAuthService(oauthConfig)
		oauthHandler = handlers.NewOAuthHandler(oauthService)
		
		// 启动定期清理过期state的任务
		oauthService.StartStateCleanupTask()
		logrus.Info("OAuth service initialized successfully")
	}

	// 创建处理器
	handler := handlers.NewHandler(cfg)

	// 创建聊天服务和处理器
	cursorService := services.NewCursorService(cfg)
	
	// Initialize ProviderRouter for multi-provider support
	// Requirements: 1.2, 1.5
	providerRouter := services.NewProviderRouter(cfg)
	
	// Register Cursor provider as fallback
	cursorProvider := services.NewCursorProvider(cursorService)
	providerRouter.RegisterProvider("cursor", cursorProvider)
	
	// Log available providers on startup
	availableProviders := providerRouter.GetAvailableProviders()
	logrus.WithFields(logrus.Fields{
		"providers": availableProviders,
		"count":     len(availableProviders),
	}).Info("Multi-provider router initialized")
	
	// Create ChatService with ProviderRouter
	chatService := services.NewChatServiceWithRouter(cursorService, providerRouter, cfg)
	chatHandler := handlers.NewChatHandlerWithRouter(chatService, providerRouter, cfg)

	// 注册路由
	setupRoutes(router, handler, cfg, oauthHandler, chatHandler)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// 启动服务器的goroutine
	go func() {
		logrus.Infof("Starting Curry2API server on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// 停止清理服务
	cleanupService.Stop()

	// 给服务器5秒时间完成处理正在进行的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func setupRoutes(router *gin.Engine, handler *handlers.Handler, cfg *config.Config, oauthHandler *handlers.OAuthHandler, chatHandler *handlers.ChatHandler) {
	// 健康检查（公开访问）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// 认证路由组（公开访问）
	auth := router.Group("/auth")
	{
		auth.POST("/send-code", handlers.SendVerificationCodeHandler) // 发送验证码
		auth.POST("/register", handlers.RegisterHandler)               // 用户注册（需要验证码）
		auth.POST("/login", handlers.LoginHandler)                     // 用户登录
		auth.POST("/logout", handlers.LogoutHandler)                   // 用户登出
		auth.GET("/me", middleware.SessionAuth(), handlers.GetCurrentUserHandler) // 获取当前用户信息
	}
	
	// OAuth 路由组（公开访问）
	if oauthHandler != nil {
		api := router.Group("/api")
		{
			oauthGroup := api.Group("/auth")
			{
				oauthGroup.GET("/:provider/login", oauthHandler.InitiateOAuthLogin)    // 发起OAuth登录
				oauthGroup.GET("/:provider/callback", oauthHandler.OAuthCallback)      // OAuth回调
			}
		}
	}

	// 用户个人设置路由组（需要会话认证）
	profile := router.Group("/profile", middleware.SessionAuth())
	{
		profile.PUT("/username", handlers.UpdateUsernameHandler) // 更新用户名
		profile.PUT("/password", handlers.UpdatePasswordHandler) // 更新密码
	}

	// API文档页面（需要会话认证）
	router.GET("/docs", middleware.SessionAuth(), handler.ServeDocs)

	// 创建 Claude Handler 实例
	claudeHandler := handlers.NewClaudeHandler(cfg)

	// API v1路由组
	v1 := router.Group("/v1")
	{
		// 模型列表
		v1.GET("/models", middleware.AuthRequired(), handler.ListModels)

		// OpenAI 聊天完成端点
		v1.POST("/chat/completions", middleware.AuthRequired(), handler.ChatCompletions)

		// Claude Messages API 端点
		v1.POST("/messages", middleware.AuthRequired(), claudeHandler.ClaudeMessages)
		v1.POST("/messages/count_tokens", middleware.AuthRequired(), claudeHandler.CountTokens)
		
		// Anthropic Responses API 端点（Codex CLI 使用）
		// Codex CLI 使用 OpenAI 格式，所以使用 ChatCompletions 处理器
		// 使用可选认证，允许没有 Authorization 头的请求
		v1.POST("/responses", middleware.OptionalAuth("sk-test-demo-2024"), handler.ChatCompletions)
	}

	// 用户公告路由组（需要会话认证）
	announcements := router.Group("/announcements", middleware.SessionAuth())
	{
		announcements.GET("", handlers.ListAnnouncementsHandler)           // 获取公告列表（包含阅读状态）
		announcements.GET("/unread-count", handlers.GetUnreadCountHandler) // 获取未读公告数量
		announcements.POST("/:id/read", handlers.MarkAsReadHandler)        // 标记公告为已读
	}

	// 用户使用统计路由组（需要会话认证）
	usage := router.Group("/api/usage", middleware.SessionAuth())
	{
		usage.GET("/stats", handlers.GetUserUsageStats)     // 获取用户使用统计
		usage.GET("/recent", handlers.GetUserRecentCalls)   // 获取最近的API调用
		usage.GET("/trends", handlers.GetUserUsageTrends)   // 获取用户使用趋势
	}

	// 用户余额路由组（需要会话认证）
	balance := router.Group("/api/balance", middleware.SessionAuth())
	{
		balance.GET("", handlers.GetBalanceHandler)                // 获取当前余额
		balance.GET("/transactions", handlers.GetTransactionsHandler) // 获取交易记录
	}

	// 用户邀请路由组（需要会话认证）
	referral := router.Group("/api/referral", middleware.SessionAuth())
	{
		referral.GET("/code", handlers.GetReferralCodeHandler)   // 获取邀请码和链接
		referral.GET("/stats", handlers.GetReferralStatsHandler) // 获取邀请统计
		referral.GET("/list", handlers.GetReferralListHandler)   // 获取邀请列表
	}

	// 模型广场路由组（需要会话认证）
	models := router.Group("/api/models", middleware.SessionAuth())
	{
		models.GET("/marketplace", handlers.GetModelMarketplaceHandler) // 获取模型广场数据
	}

	// 聊天路由组（需要会话认证）
	// Requirements: 1.1, 2.1, 3.1
	chat := router.Group("/api/chat", middleware.SessionAuth())
	{
		// 会话管理
		chat.POST("/conversations", chatHandler.CreateConversation)           // 创建会话
		chat.GET("/conversations", chatHandler.GetConversations)              // 获取会话列表
		chat.GET("/conversations/:id", chatHandler.GetConversation)           // 获取单个会话
		chat.PUT("/conversations/:id", chatHandler.UpdateConversation)        // 更新会话
		chat.DELETE("/conversations/:id", chatHandler.DeleteConversation)     // 删除会话
		chat.GET("/conversations/:id/messages", chatHandler.GetMessages)      // 获取消息列表
		chat.POST("/conversations/:id/messages", chatHandler.SendMessage)     // 发送消息(SSE)
		// 模型列表
		chat.GET("/models", chatHandler.GetModels)                            // 获取可用模型列表
	}

	// 游戏币路由组（需要会话认证）
	game := router.Group("/api/game", middleware.SessionAuth())
	{
		game.GET("/balance", handlers.GetGameBalanceHandler)           // 获取游戏币余额
		game.POST("/deduct", handlers.DeductGameCoinsHandler)          // 扣除游戏币（下注）
		game.POST("/add", handlers.AddGameCoinsHandler)                // 增加游戏币（获胜）
		game.POST("/reset", handlers.ResetGameCoinsHandler)            // 重置游戏币
		game.GET("/transactions", handlers.GetGameTransactionsHandler) // 获取游戏币交易记录
		game.POST("/migrate", handlers.MigrateLocalStorageHandler)     // 迁移 localStorage 数据

		// 游戏记录和统计路由
		game.POST("/record", handlers.CreateGameRecordHandler)         // 创建游戏记录
		game.GET("/records", handlers.GetGameRecordsHandler)           // 获取游戏记录（分页）
		game.GET("/stats", handlers.GetGameStatsHandler)               // 获取游戏统计
		game.GET("/leaderboard", handlers.GetLeaderboardHandler)       // 获取全局排行榜

		// 兑换相关路由
		game.POST("/exchange", handlers.ExchangeGameCoinsHandler)           // 游戏币兑换账户余额
		game.POST("/purchase", handlers.PurchaseGameCoinsHandler)           // 账户余额购买游戏币
		game.GET("/exchange/history", handlers.GetExchangeHistoryHandler)   // 获取兑换历史
		game.GET("/exchange/today", handlers.GetTodayExchangeAmountHandler) // 获取今日已兑换金额
	}

	// 管理路由组（需要管理员认证）
	admin := router.Group("/admin")
	admin.Use(handlers.AdminAuth())
	{
		// 密钥管理
		admin.GET("/keys", handlers.ListKeysHandler)                 // 列出所有密钥
		admin.POST("/keys", handlers.AddKeyHandler)                  // 添加新密钥
		admin.PUT("/keys/:key/toggle", handlers.ToggleKeyStatusHandler) // 切换密钥状态
		admin.PUT("/keys/:key/name", handlers.UpdateKeyNameHandler)  // 更新密钥名称
		admin.DELETE("/keys/:key", handlers.RemoveKeyHandler)        // 删除密钥

		// Cursor Session 管理
		cursorSession := admin.Group("/cursor")
		{
			cursorSession.GET("/sessions", handlers.ListCursorSessionsHandler)           // 列出所有 sessions
			cursorSession.POST("/sessions", handlers.AddCursorSessionHandler)            // 添加新 session
			cursorSession.POST("/sessions/reload", handlers.ReloadCursorSessionsHandler) // 重新加载 sessions
			cursorSession.DELETE("/sessions/:email", handlers.RemoveCursorSessionHandler) // 删除 session
			cursorSession.POST("/sessions/validate", handlers.ValidateCursorSessionHandler) // 验证 session
			cursorSession.GET("/sessions/stats", handlers.GetCursorSessionStatsHandler)  // 获取统计信息
			cursorSession.POST("/sessions/migrate-encrypt", handlers.MigrateEncryptCursorSessionsHandler) // 迁移加密数据
		}
		
		// Quota 管理
		quota := admin.Group("/quota")
		{
			quota.GET("/stats", handler.GetQuotaStats)       // 获取配额统计
			quota.PUT("/update", handler.UpdateQuotaLimit)   // 更新配额限制
			quota.POST("/reset", handler.ResetQuotas)        // 手动重置配额
		}

		// 用户管理
		admin.GET("/users", handlers.ListUsersHandler)                    // 列出所有用户
		admin.GET("/users/:id", handlers.GetUserHandler)                  // 获取用户信息
		admin.PUT("/users/:id/role", handlers.UpdateUserRoleHandler)      // 更新用户角色
		admin.PUT("/users/:id/status", handlers.ToggleUserStatusHandler)  // 启用/禁用用户
		admin.DELETE("/users/:id", handlers.DeleteUserHandler)            // 删除用户

		// 公告管理
		admin.POST("/announcements", handlers.CreateAnnouncementHandler)       // 创建公告
		admin.GET("/announcements", handlers.ListAllAnnouncementsHandler)      // 获取所有公告
		admin.DELETE("/announcements/:id", handlers.DeleteAnnouncementHandler) // 删除公告

		// 使用统计管理
		adminUsage := admin.Group("/usage")
		{
			adminUsage.GET("/stats", handlers.GetAdminUsageStats)           // 获取系统级使用统计
			adminUsage.GET("/trends", handlers.GetAdminUsageTrends)         // 获取使用趋势
			adminUsage.GET("/sessions", handlers.GetAdminCursorSessionUsage) // 获取Cursor会话使用统计
			adminUsage.GET("/export", handlers.ExportUsageData)             // 导出使用数据为CSV
			adminUsage.GET("/retention", handlers.GetRetentionConfig)       // 获取数据保留配置
			adminUsage.PUT("/retention", handlers.UpdateRetentionConfig)    // 更新数据保留期限
			adminUsage.POST("/cleanup", handlers.TriggerCleanupNow)         // 手动触发清理
			adminUsage.GET("/cleanup/stats", handlers.GetCleanupStats)      // 获取清理统计
		}

		// 余额管理
		adminBalance := admin.Group("/balance")
		{
			adminBalance.POST("/adjust", handlers.AdjustUserBalanceHandler)  // 调整用户余额
			adminBalance.GET("/users", handlers.GetAllUserBalancesHandler)   // 获取所有用户余额
		}

		// 兑换记录管理
		adminExchange := admin.Group("/exchanges")
		{
			adminExchange.GET("", handlers.AdminGetAllExchangesHandler)       // 获取所有兑换记录
			adminExchange.GET("/stats", handlers.AdminGetExchangeStatsHandler) // 获取兑换统计
		}
	}

	// 静态文件服务
	router.Static("/static", "./static")
	
	// 前端静态资源（从 dist 目录）
	router.Static("/assets", "./dist/assets")
	
	// 处理前端路由 - 所有未匹配的路由都返回 index.html
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		acceptHeader := c.GetHeader("Accept")
		
		// 检查是否是真正的API请求
		// 只有以下情况才认为是API请求：
		// 1. 明确的API路径前缀
		// 2. Accept头明确要求JSON
		isAPIRequest := false
		
		// 真正的API路径前缀检查（不包括前端路由）
		if len(path) >= 3 && path[:3] == "/v1" {
			isAPIRequest = true
		} else if len(path) >= 4 && path[:4] == "/api" {
			isAPIRequest = true
		} else if len(path) >= 5 && path[:5] == "/auth" {
			isAPIRequest = true
		} else if len(path) >= 6 && path[:6] == "/admin" {
			isAPIRequest = true
		} else if len(path) >= 7 && path[:7] == "/health" {
			isAPIRequest = true
		} else if len(path) >= 7 && path[:7] == "/static" {
			isAPIRequest = true
		} else if len(path) >= 7 && path[:7] == "/assets" {
			isAPIRequest = true
		} else if len(path) >= 8 && path[:8] == "/profile" {
			isAPIRequest = true
		} else if len(path) >= 14 && path[:14] == "/announcements" {
			isAPIRequest = true
		}
		
		// 检查Accept头是否明确要求JSON
		if !isAPIRequest && acceptHeader != "" {
			// 只有Accept头以application/json开头才认为是API请求
			if len(acceptHeader) >= 16 && acceptHeader[:16] == "application/json" {
				isAPIRequest = true
			}
		}
		
		// 如果是真正的API请求，返回JSON错误
		if isAPIRequest {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": "Endpoint not found",
					"code":    "not_found",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		
			// 对于所有其他请求（包括前端路由），返回index.html
		// 设置缓存控制头，防止浏览器缓存
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		
		// 记录前端路由请求
		logrus.WithFields(logrus.Fields{
			"path": path,
			"accept": acceptHeader,
			"user_agent": c.GetHeader("User-Agent"),
		}).Info("Serving frontend route")
		
		c.File("./dist/index.html")
	})
}
