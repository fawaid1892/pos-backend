package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// ListUsers returns paginated list of users (admin_cabang/owner only)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "admin_cabang" && role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: admin_cabang or owner only"})
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	var branchID *int64
	if bid := q.Get("branch_id"); bid != "" {
		if parsed, err := strconv.ParseInt(bid, 10, 64); err == nil {
			branchID = &parsed
		}
	}

	users, total, err := repository.ListUsers(repository.ListUsersParams{
		Page:     page,
		Limit:    limit,
		Role:     q.Get("role"),
		BranchID: branchID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.User{}
	}

	writeJSON(w, http.StatusOK, model.ListUsersResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetByID returns a single user
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	user, err := repository.FindUserByID(id)
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

// Create creates a new user (admin_cabang/owner only)
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "admin_cabang" && role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: admin_cabang or owner only"})
		return
	}

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password are required"})
		return
	}
	if req.Role == "" {
		req.Role = "kasir"
	}
	if req.Role != "admin_cabang" && req.Role != "kasir" && req.Role != "owner" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role (admin_cabang, kasir, owner)"})
		return
	}

	user, err := repository.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// Update updates a user (admin_cabang/owner only)
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "admin_cabang" && role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: admin_cabang or owner only"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Role != "" && req.Role != "admin_cabang" && req.Role != "kasir" && req.Role != "owner" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role (admin_cabang, kasir, owner)"})
		return
	}

	user, err := repository.UpdateUser(id, req)
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

// Delete soft-deletes a user (admin_cabang/owner only)
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r.Context())
	if role != "admin_cabang" && role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: admin_cabang or owner only"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := repository.SoftDeleteUser(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deactivated"})
}
