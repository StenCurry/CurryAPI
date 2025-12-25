package services

import (
	"context"
	"Curry2API-go/middleware"
	"Curry2API-go/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// authManager 处理 Cursor API 的认证逻辑
type authManager struct {
	service *CursorService
}

// fetchXIsHuman 获取 x-is-human 认证令牌
// 通过执行 Cursor 的 JavaScript 代码来获取反爬虫令牌
func (a *authManager) fetchXIsHuman(ctx context.Context) (string, error) {
	resp, err := a.service.client.R().
		SetContext(ctx).
		SetHeaders(a.scriptHeaders()).
		Get(a.service.config.ScriptURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch script: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		message := strings.TrimSpace(resp.String())
		return "", middleware.NewCursorWebError(resp.StatusCode, message)
	}

	scriptBody := resp.Bytes()
	compiled := a.prepareJS(string(scriptBody))
	value, err := utils.RunJS(compiled)
	if err != nil {
		return "", fmt.Errorf("failed to execute JS: %w", err)
	}

	logrus.WithField("length", len(value)).Debug("Fetched x-is-human token")

	return value, nil
}

// prepareJS 准备执行的 JavaScript 代码
// 将配置参数注入到 JS 脚本中
func (a *authManager) prepareJS(cursorJS string) string {
	replacer := strings.NewReplacer(
		"$$currentScriptSrc$$", a.service.config.ScriptURL,
		"$$UNMASKED_VENDOR_WEBGL$$", a.service.config.FP.UNMASKED_VENDOR_WEBGL,
		"$$UNMASKED_RENDERER_WEBGL$$", a.service.config.FP.UNMASKED_RENDERER_WEBGL,
		"$$userAgent$$", a.service.config.FP.UserAgent,
	)

	mainScript := replacer.Replace(a.service.mainJS)
	mainScript = strings.Replace(mainScript, "$$env_jscode$$", a.service.envJS, 1)
	mainScript = strings.Replace(mainScript, "$$cursor_jscode$$", cursorJS, 1)
	return mainScript
}

// scriptHeaders 返回获取脚本时需要的 HTTP 请求头
func (a *authManager) scriptHeaders() map[string]string {
	return map[string]string{
		"User-Agent":                 a.service.config.FP.UserAgent,
		"sec-ch-ua-arch":             `"x86"`,
		"sec-ch-ua-platform":         `"Windows"`,
		"sec-ch-ua":                  `"Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"`,
		"sec-ch-ua-bitness":          `"64"`,
		"sec-ch-ua-mobile":           "?0",
		"sec-ch-ua-platform-version": `"19.0.0"`,
		"sec-fetch-site":             "same-origin",
		"sec-fetch-mode":             "no-cors",
		"sec-fetch-dest":             "script",
		"referer":                    "https://cursor.com/en-US/learn/how-ai-models-work",
		"accept-language":            "zh-CN,zh;q=0.9,en;q=0.8",
	}
}
