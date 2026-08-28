package middleware

import (
	"context"
	"net/http"
	"strings"

	"checkout-api/internal/httpapi"
	"checkout-api/internal/service"
)

func AuthMiddleware(svc *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaderStr := r.Header.Get("Authorization")

		scheme := "bearer "
		if len(authHeaderStr) <= len(scheme) || !strings.EqualFold(authHeaderStr[:len(scheme)], scheme) {
			httpapi.Error(w, service.ErrInvalidCredentials)
			return
		}

		userID, err := svc.ValidateToken(authHeaderStr[len(scheme):])
		if err != nil {
			httpapi.Error(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "user", int(userID))
		next(w, r.WithContext(ctx))
	})
}
