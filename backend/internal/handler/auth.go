package handler

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "guardian-backend/pkg/httpx"
)

type AuthHandler struct {
	JWTSecret string
    AdminUsername string
    AdminPassword string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.WriteError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
        return
    }
    if req.Username != h.AdminUsername || req.Password != h.AdminPassword {
        httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
        return
    }
	claims := jwt.MapClaims{
		"sub": "1",
		"role": "admin",
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
    if err != nil {
        httpx.WriteError(w, r, http.StatusInternalServerError, "TOKEN_ERROR", "failed to sign token")
        return
    }
    httpx.WriteJSON(w, http.StatusOK, loginResponse{Token: tokenStr})
}
