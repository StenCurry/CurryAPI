package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 允许的源列表
		allowedOrigins := []string{
			"http://localhost:5173",      // 开发环境前端
			"http://localhost:8002",      // 后端
			"https://www.kesug.icu",      // 生产环境前端(www HTTPS)
			"http://www.kesug.icu",       // 生产环境前端(www HTTP)
			"https://kesug.icu",          // 生产环境前端(无www HTTPS)
			"http://kesug.icu",           // 生产环境前端(无www HTTP)
		}

		// 始终设置 CORS 头，确保所有请求都有响应
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Cache-Control, Pragma, Expires")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 检查请求来源是否在允许列表中
		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// 如果来源不在允许列表中，但是没有 Origin 头（同源请求），也允许
		if !isAllowed && origin == "" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}