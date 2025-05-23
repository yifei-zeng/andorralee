package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateUniqueID 生成唯一ID
func GenerateUniqueID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
