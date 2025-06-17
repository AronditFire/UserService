package models

import "time"

type RefreshTokenClaims struct {
	Token     string
	UserID    int64
	IssuedAt  time.Time
	ExpiresAt time.Time
}
