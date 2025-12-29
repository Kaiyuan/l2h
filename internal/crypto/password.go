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
	// Argon2 参数
	memory      = 64 * 1024 // 64 MB
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

// HashPassword 使用 Argon2 算法哈希密码
func HashPassword(password string) (string, error) {
	// 生成随机盐
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐失败: %w", err)
	}

	// 使用 Argon2id 生成哈希
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	// 编码为 base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 返回格式: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash), nil
}

// VerifyPassword 验证密码是否匹配哈希值
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// 解析哈希字符串
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("无效的哈希格式")
	}

	// 解析参数
	var version int
	var m, t, p uint32
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("解析版本失败: %w", err)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, fmt.Errorf("解析参数失败: %w", err)
	}

	// 解码盐和哈希
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("解码盐失败: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("解码哈希失败: %w", err)
	}

	// 计算密码的哈希
	otherHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(hash)))

	// 使用 constant-time 比较防止时序攻击
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// IsHashed 检查字符串是否是哈希值
func IsHashed(password string) bool {
	return strings.HasPrefix(password, "$argon2id$")
}
