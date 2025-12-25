package services

import (
	"crypto/tls"
	"Curry2API-go/config"
	"fmt"

	"gopkg.in/gomail.v2"
)

// EmailService 邮件发送服务
type EmailService struct {
	cfg *config.Config
}

// NewEmailService 创建邮件服务
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(toEmail, code string) error {
	if s.cfg.SMTPUser == "" || s.cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP configuration is not set")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "【Curry2API】邮箱验证码")

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .code-box {
            background: #f8f9fa;
            border: 2px dashed #667eea;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            margin: 30px 0;
        }
        .code {
            font-size: 32px;
            font-weight: bold;
            color: #667eea;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
        .info {
            color: #666;
            font-size: 14px;
            line-height: 1.6;
            margin: 20px 0;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #999;
            font-size: 12px;
        }
        .warning {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 12px 16px;
            margin: 20px 0;
            color: #856404;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Curry2API</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">欢迎注册 Curry2API 服务</p>
        </div>
        <div class="content">
            <p style="font-size: 16px; color: #333;">您好！</p>
            <p class="info">
                您正在注册 <strong>Curry2API</strong> 账号，请使用以下验证码完成注册：
            </p>
            <div class="code-box">
                <div class="code">%s</div>
                <p style="margin: 15px 0 0 0; color: #999; font-size: 14px;">
                    验证码有效期：<strong>10分钟</strong>
                </p>
            </div>
            <div class="warning">
                <strong>⚠️ 安全提示：</strong>请勿向任何人透露此验证码，Curry2API 工作人员不会向您索要验证码。
            </div>
            <p class="info">
                如果这不是您本人的操作，请忽略此邮件。
            </p>
        </div>
        <div class="footer">
            <p>© 2025 Curry2API - 通过 OpenAI 兼容的 API 访问 Cursor 模型</p>
            <p style="margin-top: 10px;">此邮件由系统自动发送，请勿直接回复</p>
        </div>
    </div>
</body>
</html>
`, code)

	m.SetBody("text/html", htmlBody)

	// 创建SMTP拨号器
	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)

	// 163邮箱使用SSL，需要跳过证书验证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendPasswordResetCode 发送密码重置验证码（未来扩展）
func (s *EmailService) SendPasswordResetCode(toEmail, code string) error {
	if s.cfg.SMTPUser == "" || s.cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP configuration is not set")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "【Curry2API】密码重置验证码")

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .code-box {
            background: #f8f9fa;
            border: 2px dashed #dc3545;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            margin: 30px 0;
        }
        .code {
            font-size: 32px;
            font-weight: bold;
            color: #dc3545;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
        .info {
            color: #666;
            font-size: 14px;
            line-height: 1.6;
            margin: 20px 0;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #999;
            font-size: 12px;
        }
        .warning {
            background: #f8d7da;
            border-left: 4px solid #dc3545;
            padding: 12px 16px;
            margin: 20px 0;
            color: #721c24;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 Curry2API</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">密码重置验证</p>
        </div>
        <div class="content">
            <p style="font-size: 16px; color: #333;">您好！</p>
            <p class="info">
                您正在重置 <strong>Curry2API</strong> 账号密码，请使用以下验证码：
            </p>
            <div class="code-box">
                <div class="code">%s</div>
                <p style="margin: 15px 0 0 0; color: #999; font-size: 14px;">
                    验证码有效期：<strong>10分钟</strong>
                </p>
            </div>
            <div class="warning">
                <strong>⚠️ 重要提示：</strong>如果这不是您本人的操作，说明您的账号可能存在安全风险，请立即修改密码！
            </div>
            <p class="info">
                若非本人操作，请忽略此邮件。
            </p>
        </div>
        <div class="footer">
            <p>© 2025 Curry2API - 通过 OpenAI 兼容的 API 访问 Cursor 模型</p>
            <p style="margin-top: 10px;">此邮件由系统自动发送，请勿直接回复</p>
        </div>
    </div>
</body>
</html>
`, code)

	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
