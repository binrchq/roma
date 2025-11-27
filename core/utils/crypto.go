package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"binrc.com/roma/core/global"
	"golang.org/x/crypto/bcrypt"
)

// 从配置或环境变量获取加密密钥
func getEncryptionKey() []byte {
	var key string

	// 优先从配置文件读取
	if global.CONFIG != nil && global.CONFIG.Security != nil && global.CONFIG.Security.EncryptionKey != "" {
		key = global.CONFIG.Security.EncryptionKey
	} else {
		// 从环境变量读取
		key = os.Getenv("ROMA_ENCRYPTION_KEY")
	}

	if key == "" {
		// 默认密钥（32字节，仅用于开发环境）
		// 生产环境必须在配置文件或环境变量中设置
		return []byte("roma-default-encryption-key-32bytes!!")
	}

	// 确保密钥长度为32字节（AES-256）
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// 如果密钥太短，填充到32字节
		padding := make([]byte, 32-len(keyBytes))
		keyBytes = append(keyBytes, padding...)
	} else if len(keyBytes) > 32 {
		// 如果密钥太长，截取前32字节
		keyBytes = keyBytes[:32]
	}
	return keyBytes
}

// EncryptPassword 加密密码
func EncryptPassword(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建加密块失败: %v", err)
	}

	// 使用 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 失败: %v", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %v", err)
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 Base64 编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword 解密密码
func DecryptPassword(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 检查是否是已加密的格式（Base64编码）
	// 如果不是Base64格式，可能是旧数据（明文），直接返回
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// 可能是旧数据（明文），直接返回
		return ciphertext, nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建解密块失败: %v", err)
	}

	// 使用 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 失败: %v", err)
	}

	// 提取 nonce
	nonceSize := aesGCM.NonceSize()
	if len(decoded) < nonceSize {
		// 可能是旧数据（明文），直接返回
		return ciphertext, nil
	}

	nonce, ciphertextBytes := decoded[:nonceSize], decoded[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		// 解密失败，可能是旧数据（明文），直接返回
		return ciphertext, nil
	}

	return string(plaintext), nil
}

// IsEncrypted 检查字符串是否是加密的
func IsEncrypted(text string) bool {
	if text == "" {
		return false
	}
	// 尝试 Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return false
	}
	// 检查长度是否足够包含 nonce（至少12字节）
	return len(decoded) >= 12
}

// HashPassword 使用 bcrypt 对用户密码进行不可逆加密
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %v", err)
	}
	return string(bytes), nil
}

// CheckPassword 验证用户密码（bcrypt）
func CheckPassword(hashedPassword, password string) bool {
	if hashedPassword == "" || password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
