package utils

import (
	"math/rand"
	"time"
)

func GenerateKey() string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefhiklmnorstuvwxzABCDEFGHIJKLMNOPRSTUVWXYZ0123456789" //平滑字符
	key := make([]byte, 57)
	for i := range key {
		key[i] = charset[random.Intn(len(charset))]
	}
	return string(key)
}
