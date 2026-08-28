package handlers

import (
	"checkout-api/internal/httpapi"
	"checkout-api/internal/httpapi/dto"
	"checkout-api/internal/service"
	"encoding/json"
	"net/http"
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
	if err := h.auth.SignUp(r.Context(), req.Email, req.Password);
	err != nil {
		 httpapi.Error(w, err)
		 return

	}
	w.WriteHeader(http.StatusCreated)
}


func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpapi.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.store.FindUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnprocessableEntity, ErrorMessageResponse{
				Message: "user does not exist",
			})
			return
		}
		fmt.Printf("cannot query %q", err.Error())
		httpapi.Error(w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(req.Password))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// issue jwt
	signedString, err := generateJWT(user.ID)
	if err != nil {
		fmt.Printf("cannot generate signed string %q", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// store session(refresh token)
	refreshToken, hash, err := generateRefreshToken()
	if err != nil {
		fmt.Printf("cannot generate refresh token %q", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = h.store.SaveRefreshToken(r.Context(), user.ID, hash, expiresAt)
	if err != nil {
		fmt.Printf("cannot save refresh token %q", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})

	writeJSON(w, http.StatusOK, AuthResponse{
		JWT:          signedString,
		RefreshToken: refreshToken,
	})
}

