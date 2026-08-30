package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"checkout-api/internal/repository"
	"checkout-api/internal/validation"
)

type AuthService struct {
	users      repository.UserRepo
	tokens     repository.TokenRepo
	signingKey []byte
	now        func() time.Time
}

type AuthResult struct {
	JWT          string
	RefreshToken string
	ExpiresAt    time.Time
}

func (s *AuthService) ValidateToken(accessToken string) (int64, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (any, error) {
		return s.signingKey, nil
	},
		jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return 0, fmt.Errorf("validate token: %w", ErrInvalidCredentials)
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse subject: %w", ErrInvalidCredentials)
	}

	return userID, nil
}

func (s *AuthService) SignUp(ctx context.Context, email string, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	var verr validation.ValidationError
	if !validation.IsEmail(email) || len(email) > 255 {
		verr.With("email", "email", "invalid email")
	}
	if len(password) < 12 || len(password) > 25 {
		verr.With("password", "min=12,max=25", "password must be from 12 to 25 characters")
	}
	if len(verr.Fields) != 0 {
		return &verr
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.users.CreateUser(ctx, email, hash)
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var verr validation.ValidationError
	if !validation.IsEmail(email) || len(email) > 255 {
		verr.With("email", "email", "invalid email")
	}
	if len(password) < 12 || len(password) > 25 {
		verr.With("password", "min=12,max=25", "password must be from 12 to 25 characters")
	}
	if len(verr.Fields) != 0 {
		return AuthResult{}, &verr
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("log in: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Hash, []byte(password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	signedString, err := s.generateJWT(user.ID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate jwt: %w", err)
	}

	raw, hash, err := generateRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	expiresAt := s.now().Add(7 * 24 * time.Hour)
	if err := s.tokens.SaveRefreshToken(ctx, user.ID, hash, expiresAt); err != nil {
		return AuthResult{}, fmt.Errorf("save refresh token: %w", err)
	}

	return AuthResult{JWT: signedString, RefreshToken: raw, ExpiresAt: expiresAt}, nil
}

func generateRefreshToken() (string, []byte, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", nil, err
	}
	rawToken := hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(rawToken))
	return rawToken, hash[:], nil
}

func (s *AuthService) generateJWT(userID int64) (string, error) {
	expiringTime := s.now().Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiringTime),
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(s.now()),
	})
	return token.SignedString(s.signingKey)
}
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (AuthResult, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	rt, err := s.tokens.FindRefreshToken(ctx, hash[:])
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResult{}, ErrInvalidCredentials
	}

	if err != nil {
		return AuthResult{}, fmt.Errorf("find refresh token: %w", err)
	}

	if !rt.IsActive || rt.ExpiresAt.Before(s.now()) {
		return AuthResult{}, ErrInvalidCredentials
	}

	jwtStr, err := s.generateJWT(rt.UserID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate jwt: %w", err)
	}
	raw, newHash, err := generateRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}
	expiresAt := s.now().Add(7 * 24 * time.Hour)
	if err := s.tokens.RotateRefreshToken(ctx, hash[:], newHash, rt.UserID, expiresAt); err != nil {
		return AuthResult{}, fmt.Errorf("rotate refresh token: %w", err)
	}
	return AuthResult{JWT: jwtStr, RefreshToken: raw, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := sha256.Sum256([]byte(refreshToken))
	return s.tokens.RevokeRefreshToken(ctx, hash[:])
}

func NewAuthService(users repository.UserRepo, tokens repository.TokenRepo, signingKey []byte) *AuthService {
	return &AuthService{users: users, tokens: tokens, signingKey: signingKey, now: time.Now}
}
