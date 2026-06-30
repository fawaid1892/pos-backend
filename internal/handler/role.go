package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"
)

type RoleHandler struct{}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{}
}

// RoleList returns all roles.
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	roles, err := repository.ListRoles()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if roles == nil {
		roles = []model.Role{}
	}
	writeJSON(w, http.StatusOK, roles)
}

// RoleGet returns a single role by ID.
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
		return
	}
	role, err := repository.GetRoleByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if role == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "role not found"})
		return
	}
	writeJSON(w, http.StatusOK, role)
}

// RoleCreate creates a new role.
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	role, err := repository.CreateRole(req.Name, req.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

// RoleUpdate updates a role's name/description.
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	role, err := repository.UpdateRole(id, req.Name, req.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if role == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "role not found"})
		return
	}
	writeJSON(w, http.StatusOK, role)
}

// RoleDelete soft-deletes a role.
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
		return
	}
	if err := repository.SoftDeleteRole(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}

// RolePermissions returns all permissions (global list).
func (h *RoleHandler) PermissionsList(w http.ResponseWriter, r *http.Request) {
	perms, err := repository.ListPermissions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if perms == nil {
		perms = []model.Permission{}
	}
	writeJSON(w, http.StatusOK, perms)
}

// GetRolePermissions returns all permissions assigned to a specific role.
func (h *RoleHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
		return
	}
	perms, err := repository.GetRolePermissions(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if perms == nil {
		perms = []model.Permission{}
	}
	writeJSON(w, http.StatusOK, perms)
}

// SetRolePermissions replaces all permissions assigned to a role.
func (h *RoleHandler) SetPermissions(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
		return
	}
	var req struct {
		PermissionIDs []int64 `json:"permission_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := repository.SetRolePermissions(id, req.PermissionIDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "permissions updated"})
}
