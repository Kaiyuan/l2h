package servera

import (
	"net/http"
	"strings"
)

// requireAPIKey 中间件：验证 API Key
func (s *Server) requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 Header 或 Query 参数获取 API Key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}

		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "API Key required")
			return
		}

		// 验证 API Key
		valid, err := s.db.ValidateAPIKey(apiKey)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !valid {
			writeError(w, http.StatusUnauthorized, "Invalid or expired API Key")
			return
		}

		next(w, r)
	}
}

// requireAuth 中间件：验证管理后台认证
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查 session cookie
		sessionCookie, err := r.Cookie("l2h_session")
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// 验证 session（这里简化处理，实际应该使用 JWT 或 session store）
		settings, err := s.db.GetSettings()
		if err != nil || settings == nil {
			writeError(w, http.StatusInternalServerError, "Settings not configured")
			return
		}

		// 简单的 session 验证（实际应该使用更安全的方式）
		if sessionCookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "Invalid session")
			return
		}

		next(w, r)
	}
}

// CORS 中间件
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// 从请求中提取 API Key 的辅助函数
func extractAPIKey(r *http.Request) string {
	// 优先从 Header 获取
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 从 Authorization header 获取 (Bearer token 格式)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 从 Query 参数获取
	return r.URL.Query().Get("api_key")
}
