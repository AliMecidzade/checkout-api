package service

import (
	"checkout-api/internal/repository"
	"checkout-api/internal/validation"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repository.UserRepo
	tokens     repository.TokenRepo
	signingKey []byte
	now        func() time.Time
}

type AuthResult struct {
	JWT string
	RefreshToken string
	ExpiresAt time.Time
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

	if len(password) < 12 && len(password) > 25 {
		verr.With("password", "min=12,max=25", "password must be from 12 to 25 characters")
	}

	if len(verr.Fields) == 0 {
		return &verr
	}

	hash , err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password %w" ,err)
	}

	return s.users.CreateUser(ctx, email, hash)
}

func (s *AuthService) Login(ctx context.Context, email string, password string) {

}


func NewAuthService(users repository.UserRepo, tokens repository.TokenRepo, signingKey []byte) *AuthService {
	return &AuthService{users: users, tokens: tokens, signingKey: signingKey, now: time.Now}
}
