package domain

import "time"

type User struct {
	ID    int64
	Email string
	Hash  []byte
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash []byte
	IsActive  bool
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}
