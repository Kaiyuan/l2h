// Package crypto 提供密码哈希和验证功能
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

// HashPassword 使用 Argon2id 算法对密码进行哈希
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐失败: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash), nil
}

// VerifyPassword 验证密码是否与哈希值匹配
func VerifyPassword(password, hashedPassword string) (bool, error) {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("无效的哈希格式")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("解析版本失败: %w", err)
	}

	var m, t, p uint32
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, fmt.Errorf("解析参数失败: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("解码盐失败: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("解码哈希失败: %w", err)
	}

	otherHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// IsHashed 检查字符串是否是哈希格式的密码
func IsHashed(password string) bool {
	return strings.HasPrefix(password, "$argon2id$")
}
