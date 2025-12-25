package database

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// User 用户模型
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	IsActive     bool      `json:"is_active"`
}

// CreateUser 创建新用户
func CreateUser(username, email, password, role string) (*User, error) {
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	// 插入用户
	result, err := db.Exec(
		`INSERT INTO users (username, email, password_hash, role, created_at, is_active) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		username, email, string(hashedPassword), role, time.Now(), true,
	)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	return &User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, last_login, is_active 
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, 
		&user.CreatedAt, &user.LastLogin, &user.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, last_login, is_active 
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, 
		&user.CreatedAt, &user.LastLogin, &user.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, last_login, is_active 
		 FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, 
		&user.CreatedAt, &user.LastLogin, &user.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// ListUsers 列出所有用户
func ListUsers() ([]*User, error) {
	rows, err := db.Query(
		`SELECT id, username, email, role, created_at, last_login, is_active 
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, 
			&user.CreatedAt, &user.LastLogin, &user.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// UpdateLastLogin 更新用户最后登录时间
func UpdateLastLogin(userID int64) error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE users SET last_login = ? WHERE id = ?`,
		now, userID,
	)
	return err
}

// ValidatePassword 验证密码
func ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		string(hashedPassword), userID,
	)
	return err
}

// UpdateUsername 更新用户名
func UpdateUsername(userID int64, newUsername string) error {
	_, err := db.Exec(
		`UPDATE users SET username = ? WHERE id = ?`,
		newUsername, userID,
	)
	return err
}

// DeleteUser 删除用户（软删除）
func DeleteUser(userID int64) error {
	_, err := db.Exec(
		`UPDATE users SET is_active = FALSE WHERE id = ?`,
		userID,
	)
	return err
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(userID int64, role string) error {
	_, err := db.Exec(
		`UPDATE users SET role = ? WHERE id = ?`,
		role, userID,
	)
	return err
}

// UpdateUserStatus 更新用户状态（同时更新该用户的所有API密钥状态）
func UpdateUserStatus(userID int64, isActive bool) error {
	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 更新用户状态
	_, err = tx.Exec(
		`UPDATE users SET is_active = ? WHERE id = ?`,
		isActive, userID,
	)
	if err != nil {
		return err
	}
	
	// 同时更新该用户创建的所有API密钥状态
	_, err = tx.Exec(
		`UPDATE api_keys SET is_active = ? WHERE user_id = ?`,
		isActive, userID,
	)
	if err != nil {
		return err
	}
	
	// 提交事务
	return tx.Commit()
}
