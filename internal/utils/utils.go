// Package utils 提供通用的工具函数
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

// generateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// 计算需要的字节数，base64 编码会增加约 33% 的长度
	byteLength := (length * 3) / 4
	if byteLength < length {
		byteLength = length
	}

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// WriteJSON 写入 JSON 响应
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError 写入错误响应
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// ContainsSensitiveWord 检查路径是否包含敏感词
func ContainsSensitiveWord(path string) bool {
	sensitiveWords := []string{
		"admin", "administrator", "root", "system", "config",
		"api", "internal", "private", "secret", "password",
		"login", "logout", "auth", "token", "key",
	}
	pathLower := strings.ToLower(path)
	for _, word := range sensitiveWords {
		if strings.Contains(pathLower, word) {
			return true
		}
	}
	return false
}

// ValidatePath 验证路径格式
func ValidatePath(path string) bool {
	if path == "" {
		return false
	}

	// 路径不能包含特殊字符
	invalidChars := []string{" ", "\\", "?", "#", "&", "=", "%"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	// 路径不能以 / 开头或结尾
	if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return false
	}

	return true
}

// ValidatePort 验证端口号
func ValidatePort(port int) bool {
	return port > 0 && port <= 65535
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	// 简单的邮箱格式验证
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
