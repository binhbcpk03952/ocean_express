package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JWTSecretKey = []byte("my_super_secret_key_change_in_production")

// TokenTTL là thời gian sống của token, dùng chung cho cả JWT expiry và TTL của
// session trong Redis để hai bên hết hạn đồng bộ.
const TokenTTL = 24 * time.Hour

// CustomClaims chứa thông tin custom trong payload JWT
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	HubID  string `json:"hub_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken sinh JWT token cho Employee. Trả về token và JTI (session ID)
// để tầng gọi lưu session vào Redis phục vụ thu hồi (logout).
func GenerateToken(userID, role, hubID string) (string, string, error) {
	jti := uuid.NewString()
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		HubID:  hubID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(JWTSecretKey)
	if err != nil {
		return "", "", err
	}
	return signed, jti, nil
}

// ValidateToken kiểm tra token và trả về CustomClaims
func ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
