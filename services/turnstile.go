package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// TurnstileService Cloudflare Turnstile 验证服务
type TurnstileService struct {
	secretKey string
	client    *http.Client
}

// TurnstileVerifyRequest Turnstile 验证请求
type TurnstileVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

// TurnstileVerifyResponse Turnstile 验证响应
type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

// NewTurnstileService 创建 Turnstile 服务
func NewTurnstileService(secretKey string) *TurnstileService {
	return &TurnstileService{
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyToken 验证 Turnstile token
func (s *TurnstileService) VerifyToken(token, remoteIP string) (bool, error) {
	if s.secretKey == "" {
		logrus.Error("Turnstile secret key not configured, verification required")
		return false, fmt.Errorf("turnstile verification is required but not configured")
	}

	if token == "" {
		logrus.Warn("Empty Turnstile token provided")
		return false, fmt.Errorf("turnstile token is required")
	}

	reqBody := TurnstileVerifyRequest{
		Secret:   s.secretKey,
		Response: token,
		RemoteIP: remoteIP,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return false, fmt.Errorf("failed to send verification request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	var verifyResp TurnstileVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !verifyResp.Success {
		logrus.Warnf("Turnstile verification failed: %v", verifyResp.ErrorCodes)
		return false, fmt.Errorf("verification failed: %v", verifyResp.ErrorCodes)
	}

	logrus.Infof("Turnstile verification successful for IP: %s", remoteIP)
	return true, nil
}
