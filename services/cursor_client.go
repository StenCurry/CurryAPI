package services

import (
	"context"
	"Curry2API-go/middleware"
	"Curry2API-go/utils"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// httpClient 处理与 Cursor API 的 HTTP 通信
type httpClient struct {
	service *CursorService
}

// chatHeaders 返回聊天请求需要的 HTTP 请求头
// 包含浏览器指纹伪装和反爬虫令牌
func (h *httpClient) chatHeaders(xIsHuman string) map[string]string {
	return map[string]string{
		"User-Agent":                 h.service.config.FP.UserAgent,
		"Content-Type":               "application/json",
		"sec-ch-ua-platform":         `"Windows"`,
		"x-path":                     "/api/chat",
		"sec-ch-ua":                  `"Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"`,
		"x-method":                   "POST",
		"sec-ch-ua-bitness":          `"64"`,
		"sec-ch-ua-mobile":           "?0",
		"sec-ch-ua-arch":             `"x86"`,
		"x-is-human":                 xIsHuman,
		"sec-ch-ua-platform-version": `"19.0.0"`,
		"origin":                     "https://cursor.com",
		"sec-fetch-site":             "same-origin",
		"sec-fetch-mode":             "cors",
		"sec-fetch-dest":             "empty",
		"referer":                    "https://cursor.com/en-US/learn/how-ai-models-work",
		"accept-language":            "zh-CN,zh;q=0.9,en;q=0.8",
		"priority":                   "u=1, i",
	}
}

// sendChatRequest 发送聊天请求到 Cursor API
// 优先使用 Cursor session，失败时回退到 x-is-human 方式
// 返回 HTTP 响应和使用的 session，调用者负责处理响应流
func (h *httpClient) sendChatRequest(ctx context.Context, xIsHuman string, jsonPayload []byte) (*http.Response, *middleware.CursorSessionInfo, error) {
	sessionMgr := middleware.GetCursorSessionManager()

	// 1. 尝试使用 Cursor session（如果有）
	if sessionMgr.HasValidSessions() {
		session, err := sessionMgr.GetValidSession()
		if err == nil {
			resp, err := h.sendWithSession(ctx, session, jsonPayload)
			if err == nil && resp.StatusCode == http.StatusOK {
				// Session 成功
				sessionMgr.MarkSessionSuccess(session)
				logrus.Debugf("Request sent using Cursor session: %s", session.Email)
				return resp, session, nil
			}

			// Session 失败，记录详细错误信息
			var failReason string
			var respBody string
			if err != nil {
				failReason = fmt.Sprintf("request error: %v", err)
			} else if resp != nil {
				failReason = fmt.Sprintf("status code: %d", resp.StatusCode)
				// 读取响应体以获取更多错误信息
				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					respBody = string(body)
					if len(respBody) > 200 {
						respBody = respBody[:200] + "..."
					}
				}
			} else {
				failReason = "unknown error"
			}
			
			if resp != nil {
				resp.Body.Close()
			}
			sessionMgr.MarkSessionFailed(session)
			logFields := logrus.Fields{
				"session": session.Email,
				"reason":  failReason,
			}
			if respBody != "" {
				logFields["response"] = respBody
			}
			logrus.WithFields(logFields).Warn("Cursor session failed, falling back to x-is-human")
		}
	}

	// 2. 回退到 x-is-human 方式
	logrus.Debug("Using x-is-human fallback method")
	resp, err := h.sendWithXIsHuman(ctx, xIsHuman, jsonPayload)
	return resp, nil, err
}

// sendWithSession 使用 Cursor session 发送请求
func (h *httpClient) sendWithSession(ctx context.Context, session *middleware.CursorSessionInfo, jsonPayload []byte) (*http.Response, error) {
	// 构建 Cookie 字符串 - 使用正确的 Cookie 名称
	cookies := fmt.Sprintf("WorkosCursorSessionToken=%s", session.Token)
	if len(session.ExtraCookies) > 0 {
		var extraCookies []string
		for name, value := range session.ExtraCookies {
			extraCookies = append(extraCookies, fmt.Sprintf("%s=%s", name, value))
		}
		cookies = cookies + "; " + strings.Join(extraCookies, "; ")
	}

	// 使用与 x-is-human 相同的 User-Agent
	userAgent := session.UserAgent
	if userAgent == "" {
		userAgent = h.service.config.FP.UserAgent
	}

	// 发送请求 - 使用 Cookie 和 Authorization 两种方式
	resp, err := h.service.client.R().
		SetContext(ctx).
		SetHeader("User-Agent", userAgent).
		SetHeader("Content-Type", "application/json").
		SetHeader("Cookie", cookies).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", session.Token)).
		SetHeader("sec-ch-ua-platform", `"Windows"`).
		SetHeader("x-path", "/api/chat").
		SetHeader("sec-ch-ua", `"Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"`).
		SetHeader("x-method", "POST").
		SetHeader("sec-ch-ua-bitness", `"64"`).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-arch", `"x86"`).
		SetHeader("sec-ch-ua-platform-version", `"19.0.0"`).
		SetHeader("Origin", "https://cursor.com").
		SetHeader("sec-fetch-site", "same-origin").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("Referer", "https://cursor.com/en-US/learn/how-ai-models-work").
		SetHeader("accept-language", "zh-CN,zh;q=0.9,en;q=0.8").
		SetHeader("priority", "u=1, i").
		SetBody(jsonPayload).
		DisableAutoReadResponse().
		Post(cursorAPIURL)
	if err != nil {
		return nil, err
	}

	return resp.Response, nil
}

// sendWithXIsHuman 使用 x-is-human 方式发送请求（原有方法）
func (h *httpClient) sendWithXIsHuman(ctx context.Context, xIsHuman string, jsonPayload []byte) (*http.Response, error) {
	resp, err := h.service.client.R().
		SetContext(ctx).
		SetHeaders(h.chatHeaders(xIsHuman)).
		SetBody(jsonPayload).
		DisableAutoReadResponse().
		Post(cursorAPIURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Response.Body)
		resp.Response.Body.Close()
		message := strings.TrimSpace(string(body))
		if strings.Contains(message, "Attention Required! | Cloudflare") {
			message = "Cloudflare 403"
		}
		return nil, middleware.NewCursorWebError(resp.StatusCode, message)
	}

	return resp.Response, nil
}

// consumeSSE 消费 Server-Sent Events 流
// 将 SSE 数据转换为 channel 发送
func (h *httpClient) consumeSSE(ctx context.Context, resp *http.Response, output chan interface{}, session *middleware.CursorSessionInfo) {
	defer close(output)

	if err := utils.ReadSSEStream(ctx, resp, output); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		errResp := middleware.NewCursorWebError(http.StatusBadGateway, err.Error())
		select {
		case output <- errResp:
		default:
			logrus.WithError(err).Warn("failed to push SSE error to channel")
		}
	}
}
