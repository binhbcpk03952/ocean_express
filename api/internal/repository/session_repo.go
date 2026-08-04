package repository

import (
	"context"
	"ocean-express-api/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) domain.SessionRepository {
	return &sessionRepository{rdb: rdb}
}

// key sinh khóa Redis cho một phiên. Prefix để tách namespace với dữ liệu khác.
func sessionKey(jti string) string {
	return "session:" + jti
}

func (r *sessionRepository) Create(ctx context.Context, jti, userID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, sessionKey(jti), userID, ttl).Err()
}

func (r *sessionRepository) Exists(ctx context.Context, jti string) (bool, error) {
	n, err := r.rdb.Exists(ctx, sessionKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, jti string) error {
	return r.rdb.Del(ctx, sessionKey(jti)).Err()
}

func otpKey(identifier string) string {
	return "otp:" + identifier
}

func (r *sessionRepository) SetOTP(ctx context.Context, identifier, otp string, ttl time.Duration) error {
	return r.rdb.Set(ctx, otpKey(identifier), otp, ttl).Err()
}

func (r *sessionRepository) GetOTP(ctx context.Context, identifier string) (string, error) {
	val, err := r.rdb.Get(ctx, otpKey(identifier)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *sessionRepository) DeleteOTP(ctx context.Context, identifier string) error {
	return r.rdb.Del(ctx, otpKey(identifier)).Err()
}
