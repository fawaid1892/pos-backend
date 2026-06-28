package handler

import (
	"encoding/json"
	"net/http"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

type BranchHandler struct{}

func NewBranchHandler() *BranchHandler {
	return &BranchHandler{}
}

// List godoc
// @Summary      List branches
// @Description  Get all active branches
// @Tags         Branches
// @Produce      json
// @Success      200  {array}   model.Branch
// @Security     BearerAuth
// @Router       /branches [get]
func (h *BranchHandler) List(w http.ResponseWriter, r *http.Request) {
	branches, err := repository.ListBranches(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if branches == nil {
		branches = []model.Branch{}
	}
	writeJSON(w, http.StatusOK, branches)
}

// GetByID godoc
// @Summary      Get branch by ID
// @Description  Get a single branch by its UUID
// @Tags         Branches
// @Produce      json
// @Param id   path      string  true  "Branch UUID"
// @Success      200  {object}  model.Branch
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /branches/{id} [get]
func (h *BranchHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	branch, err := repository.GetBranchByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if branch == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "branch not found"})
		return
	}
	writeJSON(w, http.StatusOK, branch)
}

// Create godoc
// @Summary      Create a branch
// @Description  Create a new branch
// @Tags         Branches
// @Accept       json
// @Produce      json
// @Param request body model.CreateBranchRequest true "Branch data"
// @Success      201  {object}  model.Branch
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /branches [post]
func (h *BranchHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	branch, err := repository.CreateBranch(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, branch)
}

// Update godoc
// @Summary      Update a branch
// @Description  Update branch details by UUID
// @Tags         Branches
// @Accept       json
// @Produce      json
// @Param id       path      string                    true  "Branch UUID"
// @Param request  body      model.UpdateBranchRequest  true  "Updated branch data"
// @Success      200  {object}  model.Branch
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /branches/{id} [put]
func (h *BranchHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req model.UpdateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	branch, err := repository.UpdateBranch(r.Context(), id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if branch == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "branch not found"})
		return
	}
	writeJSON(w, http.StatusOK, branch)
}

// Delete godoc
// @Summary      Soft-delete a branch
// @Description  Soft-delete (deactivate) a branch by UUID
// @Tags         Branches
// @Produce      json
// @Param id   path      string  true  "Branch UUID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /branches/{id} [delete]
func (h *BranchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := repository.SoftDeleteBranch(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
