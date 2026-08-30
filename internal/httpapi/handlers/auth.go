package handlers

import (
	"checkout-api/internal/httpapi"
	"checkout-api/internal/httpapi/dto"
	"checkout-api/internal/service"
	"encoding/json"
	"net/http"
	"time"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.BadRequest(w, "invalid request body")
		return
	}
	if err := h.auth.SignUp(r.Context(), req.Email, req.Password); err != nil {
		httpapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	givenRefreshToken := r.URL.Query().Get("refresh_token")
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		givenRefreshToken = cookie.Value
	}

	if givenRefreshToken == "" {
		httpapi.BadRequest(w, "missing refresh token")
		return
	}

	res, err := h.auth.RefreshToken(r.Context(), givenRefreshToken)

	if err != nil {
		httpapi.Error(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(res.ExpiresAt).Seconds()),
	})

	httpapi.JSON(w, http.StatusOK, dto.AuthResponse{JWT: res.JWT, RefreshToken: res.RefreshToken})

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token", Value: "", Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.BadRequest(w, "invalid request body")
		return
	}

	res, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(res.ExpiresAt).Seconds()),
	})

	httpapi.JSON(w, http.StatusOK, dto.AuthResponse{
		JWT:          res.JWT,
		RefreshToken: res.RefreshToken,
	})
}
