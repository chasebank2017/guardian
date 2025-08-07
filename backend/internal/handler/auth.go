package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	JWTSecret string
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Username != "admin" || req.Password != "password" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(loginResponse{Token: tokenStr})
}
