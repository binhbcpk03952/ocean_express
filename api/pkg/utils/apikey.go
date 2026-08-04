package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateAPIKey sinh ra một API key ngẫu nhiên an toàn cho Shop.
// Định dạng: "oe_" + 32 ký tự hex (16 bytes ngẫu nhiên).
func GenerateAPIKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "oe_" + hex.EncodeToString(b), nil
}
