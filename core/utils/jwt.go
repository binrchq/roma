package utils

import (
	"errors"
	"os"
	"time"

	"binrc.com/roma/core/global"
	"github.com/golang-jwt/jwt/v5"
)

// getJWTSecret 从配置或环境变量获取 JWT 密钥
func getJWTSecret() []byte {
	var secret string

	// 优先从配置文件读取
	if global.CONFIG != nil && global.CONFIG.Security != nil && global.CONFIG.Security.JWT != nil && global.CONFIG.Security.JWT.Secret != "" {
		secret = global.CONFIG.Security.JWT.Secret
	} else {
		// 从环境变量读取
		secret = os.Getenv("ROMA_JWT_SECRET")
	}

	if secret == "" {
		// 默认密钥（仅用于开发环境）
		// 生产环境必须在配置文件或环境变量中设置
		return []byte("06b79d28cdf03a575012a36f36d0ee738806b05072548efeca029a5ee1de85a9")
	}

	return []byte(secret)
}

// getJWTExpireHours 获取 JWT 过期时间（小时）
func getJWTExpireHours() int {
	if global.CONFIG != nil && global.CONFIG.Security != nil && global.CONFIG.Security.JWT != nil && global.CONFIG.Security.JWT.ExpireHours > 0 {
		return global.CONFIG.Security.JWT.ExpireHours
	}
	// 默认24小时
	return 24
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT token
func GenerateJWT(userID uint, username string) (string, error) {
	expireHours := getJWTExpireHours()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ParseJWT 解析 JWT token
func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
