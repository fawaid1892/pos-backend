package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"pos-multi-branch/backend/internal/config"
	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password required"})
		return
	}

	user, err := repository.FindUserByUsername(req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if user == nil || !repository.VerifyPassword(user.Password, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Role, user.BranchID, h.cfg.JWTExpiryHours)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}

	// Generate refresh token
	refreshTokenStr, err := repository.GenerateRefreshTokenString()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh token generation failed"})
		return
	}
	refreshExpiry := time.Now().Add(time.Duration(h.cfg.JWTRefreshExpiryHours) * time.Hour)
	if err := repository.CreateRefreshToken(user.ID, refreshTokenStr, refreshExpiry); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh token storage failed"})
		return
	}

	writeJSON(w, http.StatusOK, model.LoginResponse{
		Token:        token,
		RefreshToken: refreshTokenStr,
		User:         *user,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token required"})
		return
	}

	// Validate refresh token
	rt, err := repository.FindRefreshToken(req.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if rt == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		return
	}

	// Lookup the user
	user, err := repository.FindUserByID(rt.UserID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	// Revoke the old refresh token (rotation)
	if err := repository.RevokeRefreshToken(req.RefreshToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke old refresh token"})
		return
	}

	// Generate new access token
	token, err := middleware.GenerateToken(user.ID, user.Role, user.BranchID, h.cfg.JWTExpiryHours)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}

	// Generate new refresh token
	newRefreshTokenStr, err := repository.GenerateRefreshTokenString()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh token generation failed"})
		return
	}
	refreshExpiry := time.Now().Add(time.Duration(h.cfg.JWTRefreshExpiryHours) * time.Hour)
	if err := repository.CreateRefreshToken(user.ID, newRefreshTokenStr, refreshExpiry); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh token storage failed"})
		return
	}

	writeJSON(w, http.StatusOK, model.RefreshTokenResponse{
		Token:        token,
		RefreshToken: newRefreshTokenStr,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := repository.FindUserByID(userID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, model.MeResponse{User: *user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Stateless JWT — client must discard token.
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
