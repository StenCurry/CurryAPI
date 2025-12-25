package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"Curry2API-go/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var db *sql.DB

// Init 初始化数据库连接
func Init(cfg *config.Config) error {
	var err error
	
	// 构建 MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.MySQLUser,
		cfg.MySQLPassword,
		cfg.MySQLHost,
		cfg.MySQLPort,
		cfg.MySQLDatabase,
	)
	
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	logrus.Info("Database connected successfully")
	
	// Fix any tables with incompatible foreign key types before creating tables
	fixIncompatibleTables()
	
	// 创建表
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	return nil
}

// fixIncompatibleTables fixes tables that may have been created with incompatible foreign key types
func fixIncompatibleTables() {
	// First, check if users table has INT id instead of BIGINT
	var usersIdType string
	err := db.QueryRow(`
		SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'users' 
		AND COLUMN_NAME = 'id'
	`).Scan(&usersIdType)
	
	if err == nil && !strings.Contains(strings.ToLower(usersIdType), "bigint") {
		logrus.Infof("Users table has incompatible id type (%s), need to fix all dependent tables...", usersIdType)
		
		// Drop all tables that have foreign keys to users in reverse dependency order
		tablesToDrop := []string{
			"chat_messages",
			"chat_conversations",
			"announcement_reads",
			"announcements",
			"oauth_accounts",
			"oauth_states",
			"game_records",
			"exchange_records",
			"game_coin_transactions",
			"user_game_balances",
			"referrals",
			"balance_transactions",
			"user_balances",
			"usage_records",
			"verification_codes",
			"sessions",
			"api_keys",
			"cursor_sessions",
			"users",
		}
		
		for _, table := range tablesToDrop {
			_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		}
		
		logrus.Info("All tables dropped for recreation with correct schema")
		return
	}
	
	// List of tables that reference users(id) and need BIGINT user_id
	tablesToCheck := []struct {
		tableName  string
		columnName string
		childTable string // child table that needs to be dropped first (if any)
	}{
		{"chat_conversations", "user_id", "chat_messages"},
		{"announcements", "created_by", "announcement_reads"},
		{"announcement_reads", "user_id", ""},
		{"oauth_accounts", "user_id", ""},
		{"user_game_balances", "user_id", ""},
		{"game_coin_transactions", "user_id", ""},
		{"exchange_records", "user_id", ""},
		{"game_records", "user_id", ""},
	}
	
	for _, table := range tablesToCheck {
		var columnType string
		err := db.QueryRow(`
			SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = ? 
			AND COLUMN_NAME = ?
		`, table.tableName, table.columnName).Scan(&columnType)
		
		if err != nil {
			// Table doesn't exist or column doesn't exist, nothing to fix
			continue
		}
		
		// If column is not BIGINT, we need to recreate the table
		if !strings.Contains(strings.ToLower(columnType), "bigint") {
			logrus.Infof("Fixing table %s with incompatible %s type (%s)...", table.tableName, table.columnName, columnType)
			
			// Drop child table first if exists
			if table.childTable != "" {
				_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table.childTable))
			}
			// Drop the table
			_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table.tableName))
			
			logrus.Infof("Table %s dropped for recreation with correct schema", table.tableName)
		}
	}
}

// GetDB 获取数据库连接
func GetDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db, nil
}

// createTables 创建所有必要的表
func createTables() error {
	tables := []string{
		// 用户表
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(32) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			INDEX idx_username (username),
			INDEX idx_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// API密钥表
		`CREATE TABLE IF NOT EXISTS api_keys (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			key_value VARCHAR(255) NOT NULL UNIQUE,
			masked_key VARCHAR(255) NOT NULL,
			token_name VARCHAR(255) COMMENT 'Optional descriptive name for the token',
			user_id BIGINT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			usage_count BIGINT NOT NULL DEFAULT 0,
			last_used_at DATETIME COMMENT 'Last time this token was used',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			INDEX idx_key (key_value),
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// Cursor Session表
		`CREATE TABLE IF NOT EXISTS cursor_sessions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			token TEXT NOT NULL,
			user_agent VARCHAR(500),
			extra_cookies TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_used DATETIME,
			last_check DATETIME,
			expires_at DATETIME,
			is_valid BOOLEAN NOT NULL DEFAULT TRUE,
			usage_count BIGINT NOT NULL DEFAULT 0,
			fail_count INT NOT NULL DEFAULT 0,
			INDEX idx_email (email),
			INDEX idx_is_valid (is_valid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// 用户会话表
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(64) PRIMARY KEY,
			user_id BIGINT NOT NULL,
			username VARCHAR(32) NOT NULL,
			role VARCHAR(20) NOT NULL,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// 验证码表
		`CREATE TABLE IF NOT EXISTS verification_codes (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			code VARCHAR(6) NOT NULL,
			code_type VARCHAR(20) NOT NULL,
			ip_address VARCHAR(45),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			used BOOLEAN NOT NULL DEFAULT FALSE,
			INDEX idx_email_type (email, code_type),
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// 公告表
		`CREATE TABLE IF NOT EXISTS announcements (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			created_by BIGINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			INDEX idx_created_at (created_at),
			INDEX idx_is_active (is_active),
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// 公告阅读记录表
		`CREATE TABLE IF NOT EXISTS announcement_reads (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			announcement_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			read_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_announcement_user (announcement_id, user_id),
			INDEX idx_user_id (user_id),
			INDEX idx_announcement_id (announcement_id),
			FOREIGN KEY (announcement_id) REFERENCES announcements(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// OAuth账号关联表
		`CREATE TABLE IF NOT EXISTS oauth_accounts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			provider VARCHAR(50) NOT NULL COMMENT 'OAuth provider: google, github',
			provider_user_id VARCHAR(255) NOT NULL COMMENT 'User ID from OAuth provider',
			email VARCHAR(255) COMMENT 'Email from OAuth provider',
			username VARCHAR(255) COMMENT 'Username from OAuth provider',
			avatar_url VARCHAR(500) COMMENT 'Avatar URL from OAuth provider',
			access_token TEXT COMMENT 'Encrypted access token',
			refresh_token TEXT COMMENT 'Encrypted refresh token',
			token_expires_at DATETIME COMMENT 'Token expiration time',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_provider_user (provider, provider_user_id),
			INDEX idx_oauth_user_id (user_id),
			INDEX idx_oauth_provider (provider),
			INDEX idx_oauth_email (email),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		// OAuth状态令牌表
		`CREATE TABLE IF NOT EXISTS oauth_states (
			state VARCHAR(64) PRIMARY KEY COMMENT 'Random state token for CSRF protection',
			provider VARCHAR(50) NOT NULL COMMENT 'OAuth provider: google, github',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL COMMENT 'State expiration time (10 minutes)',
			INDEX idx_oauth_states_expires (expires_at),
			INDEX idx_oauth_states_provider (provider)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// API使用记录表
		`CREATE TABLE IF NOT EXISTS usage_records (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			username VARCHAR(100) NOT NULL,
			api_token VARCHAR(255) NOT NULL,
			token_name VARCHAR(255) COMMENT 'Token name at time of request',
			model VARCHAR(100) NOT NULL,
			prompt_tokens INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0,
			total_tokens INT NOT NULL DEFAULT 0,
			cursor_session VARCHAR(255) COMMENT 'Cursor session email used',
			status_code INT NOT NULL,
			error_message TEXT,
			request_time DATETIME NOT NULL,
			response_time DATETIME NOT NULL,
			duration_ms INT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_time (user_id, request_time DESC),
			INDEX idx_token_time (api_token, request_time DESC),
			INDEX idx_model_time (model, request_time DESC),
			INDEX idx_request_time (request_time DESC)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 用户余额表 (User Balance System)
		`CREATE TABLE IF NOT EXISTS user_balances (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE,
			balance DECIMAL(10, 6) NOT NULL DEFAULT 50.000000 COMMENT 'Balance in USD',
			status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'active or exhausted',
			referral_code VARCHAR(6) NOT NULL UNIQUE COMMENT 'Unique 6-character referral code',
			total_consumed DECIMAL(10, 6) NOT NULL DEFAULT 0 COMMENT 'Total consumed amount',
			total_recharged DECIMAL(10, 6) NOT NULL DEFAULT 50.000000 COMMENT 'Total recharged amount including initial',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_balances_status (status),
			INDEX idx_user_balances_referral_code (referral_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 余额交易记录表 (Balance Transactions)
		`CREATE TABLE IF NOT EXISTS balance_transactions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			type VARCHAR(30) NOT NULL COMMENT 'initial, api_usage, referral_bonus, admin_adjust',
			amount DECIMAL(10, 6) NOT NULL COMMENT 'Positive for credit, negative for debit',
			balance_after DECIMAL(10, 6) NOT NULL COMMENT 'Balance after this transaction',
			tokens INT DEFAULT 0 COMMENT 'Token count for API usage',
			description VARCHAR(500),
			related_user_id BIGINT COMMENT 'Related user ID for referral',
			admin_id BIGINT COMMENT 'Admin ID for admin adjustments',
			api_token VARCHAR(255) COMMENT 'API token used for API usage',
			model VARCHAR(100) COMMENT 'Model used for API usage',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_transactions_user_time (user_id, created_at DESC),
			INDEX idx_transactions_type (type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 邀请关系表 (Referrals)
		`CREATE TABLE IF NOT EXISTS referrals (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			referrer_id BIGINT NOT NULL COMMENT 'User who referred',
			referee_id BIGINT NOT NULL UNIQUE COMMENT 'User who was referred',
			bonus_amount DECIMAL(10, 6) NOT NULL DEFAULT 50.000000 COMMENT 'Bonus amount awarded',
			status VARCHAR(20) NOT NULL DEFAULT 'completed',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_referrals_referrer (referrer_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 用户游戏币余额表 (User Game Balances)
		`CREATE TABLE IF NOT EXISTS user_game_balances (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE,
			balance DECIMAL(10, 2) NOT NULL DEFAULT 100.00 COMMENT 'Game coin balance',
			total_won DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT 'Total coins won from games',
			total_lost DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT 'Total coins lost in games',
			total_exchanged DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT 'Total coins exchanged to balance',
			games_played INT NOT NULL DEFAULT 0 COMMENT 'Total games played',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_game_balances_user_id (user_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 游戏币交易记录表 (Game Coin Transactions)
		`CREATE TABLE IF NOT EXISTS game_coin_transactions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			type VARCHAR(30) NOT NULL COMMENT 'initial, game_bet, game_win, exchange, reset',
			game_type VARCHAR(30) COMMENT 'wheel, coin, number',
			amount DECIMAL(10, 2) NOT NULL COMMENT 'Positive for credit, negative for debit',
			balance_after DECIMAL(10, 2) NOT NULL COMMENT 'Balance after this transaction',
			description VARCHAR(500),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_game_transactions_user_time (user_id, created_at DESC),
			INDEX idx_game_transactions_type (type),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 兑换记录表 (Exchange Records)
		`CREATE TABLE IF NOT EXISTS exchange_records (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			game_coins_amount DECIMAL(10, 2) NOT NULL COMMENT 'Game coins exchanged',
			usd_amount DECIMAL(10, 6) NOT NULL COMMENT 'USD amount received',
			exchange_rate DECIMAL(10, 4) NOT NULL DEFAULT 1.0000 COMMENT 'Exchange rate applied',
			status VARCHAR(20) NOT NULL DEFAULT 'completed' COMMENT 'completed, failed',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_exchange_records_user_time (user_id, created_at DESC),
			INDEX idx_exchange_records_date (created_at),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 游戏记录表 (Game Records)
		`CREATE TABLE IF NOT EXISTS game_records (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			game_type VARCHAR(30) NOT NULL COMMENT 'wheel, coin, number',
			bet_amount DECIMAL(10, 2) NOT NULL,
			result VARCHAR(10) NOT NULL COMMENT 'win, lose',
			payout DECIMAL(10, 2) NOT NULL DEFAULT 0,
			net_profit DECIMAL(10, 2) NOT NULL COMMENT 'payout - bet_amount',
			details JSON COMMENT 'Game-specific details',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_game_records_user_time (user_id, created_at DESC),
			INDEX idx_game_records_type (game_type),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 聊天会话表 (Chat Conversations)
		`CREATE TABLE IF NOT EXISTS chat_conversations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL DEFAULT '新对话',
			model VARCHAR(100) NOT NULL,
			system_prompt TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_updated (user_id, updated_at DESC),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 聊天消息表 (Chat Messages)
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			conversation_id BIGINT NOT NULL,
			role ENUM('user', 'assistant', 'system') NOT NULL,
			content MEDIUMTEXT NOT NULL,
			tokens INT DEFAULT 0,
			cost DECIMAL(10,6) DEFAULT 0.000000,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_conversation_created (conversation_id, created_at),
			FOREIGN KEY (conversation_id) REFERENCES chat_conversations(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	
	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	
	logrus.Info("All database tables created successfully")
	
	// Run migrations for existing tables
	if err := runMigrations(); err != nil {
		logrus.Warnf("Some migrations failed (may be expected if columns already exist): %v", err)
	}
	
	return nil
}

// runMigrations runs schema migrations for existing tables
func runMigrations() error {
	migrations := []string{
		// Add token_name column to api_keys if not exists
		`ALTER TABLE api_keys ADD COLUMN token_name VARCHAR(255) COMMENT 'Optional descriptive name for the token' AFTER masked_key`,
		// Add last_used_at column to api_keys if not exists
		`ALTER TABLE api_keys ADD COLUMN last_used_at DATETIME COMMENT 'Last time this token was used' AFTER usage_count`,
		// Add quota_limit column to api_keys for token spending limits
		`ALTER TABLE api_keys ADD COLUMN quota_limit DECIMAL(10, 6) DEFAULT NULL COMMENT 'Quota limit in USD, NULL means unlimited'`,
		// Add quota_used column to api_keys for tracking consumed quota
		`ALTER TABLE api_keys ADD COLUMN quota_used DECIMAL(10, 6) DEFAULT 0 COMMENT 'Quota used in USD'`,
		// Add expires_at column to api_keys for token expiration
		`ALTER TABLE api_keys ADD COLUMN expires_at DATETIME DEFAULT NULL COMMENT 'Expiration time, NULL means never expires'`,
		// Add allowed_models column to api_keys for model restrictions
		`ALTER TABLE api_keys ADD COLUMN allowed_models TEXT DEFAULT NULL COMMENT 'JSON array of allowed models, NULL means all models'`,
		// Add wins column to user_game_balances for tracking win count
		`ALTER TABLE user_game_balances ADD COLUMN wins INT NOT NULL DEFAULT 0 COMMENT 'Total wins' AFTER games_played`,
	}
	
	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Ignore "Duplicate column name" errors - column already exists
			if !isDuplicateColumnError(err) {
				logrus.Warnf("Migration warning: %v", err)
			}
		}
	}
	
	logrus.Info("Database migrations completed")
	return nil
}



// isDuplicateColumnError checks if the error is a duplicate column error
func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate column name") || strings.Contains(errStr, "1060")
}
