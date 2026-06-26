package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	branchID := middleware.GetBranchID(r.Context())

	// Owner sees all users (including owners from other branches).
	// Admin/kasir only sees users in their own branch + owners.
	if role == "owner" {
		users, err := repository.ListUsers(r.Context(), nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if users == nil {
			users = []model.User{}
		}
		writeJSON(w, http.StatusOK, model.UserListResponse{Users: users, Total: len(users)})
		return
	}

	// Admin/kasir — only their branch + owners
	users, err := repository.ListUsers(r.Context(), branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.User{}
	}
	writeJSON(w, http.StatusOK, model.UserListResponse{Users: users, Total: len(users)})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only owner can create users"})
		return
	}

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username is required"})
		return
	}
	if req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password is required"})
		return
	}
	if req.Role == "" {
		req.Role = "kasir"
	}
	if req.Role != "admin" && req.Role != "kasir" && req.Role != "owner" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be admin, kasir, or owner"})
		return
	}

	user, err := repository.CreateUser(r.Context(), req)
	if err != nil {
		// Detect unique constraint violation (duplicate username)
		if isPGUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "username already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only owner can update users"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := repository.UpdateUser(r.Context(), id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// isPGUniqueViolation checks if the error is a PostgreSQL unique constraint violation (code 23505).
func isPGUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if ok := errors.As(err, &pgErr); ok {
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "23505")
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only owner can deactivate users"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	// Prevent self-deactivation
	userID := middleware.GetUserID(r.Context())
	if userID == id {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot deactivate yourself"})
		return
	}

	if err := repository.SoftDeleteUser(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deactivated"})
}
