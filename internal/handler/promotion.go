package handler

import (
	"encoding/json"
	"net/http"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

type PromotionHandler struct{}

func NewPromotionHandler() *PromotionHandler {
	return &PromotionHandler{}
}

func (h *PromotionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var params repository.ListPromotionsParams

	if v := q.Get("type"); v != "" {
		params.Type = v
	}
	if v := q.Get("branch_id"); v != "" {
		bid, err := uuid.Parse(v)
		if err == nil {
			params.BranchID = &bid
		}
	}
	if v := q.Get("active"); v == "true" {
		t := true
		params.Active = &t
	}
	if v := q.Get("expired"); v == "true" {
		t := true
		params.Expired = &t
	}

	promos, err := repository.ListPromotions(params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if promos == nil {
		promos = []model.Promotion{}
	}
	writeJSON(w, http.StatusOK, promos)
}

func (h *PromotionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if req.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type is required"})
		return
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start_date and end_date are required"})
		return
	}

	// Default scope
	if req.Scope == "" {
		req.Scope = "selected"
	}

	promo, err := repository.CreatePromotion(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, promo)
}

func (h *PromotionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	promo, err := repository.GetPromotionByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if promo == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "promotion not found"})
		return
	}
	writeJSON(w, http.StatusOK, promo)
}

func (h *PromotionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req model.UpdatePromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	promo, err := repository.UpdatePromotion(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if promo == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "promotion not found"})
		return
	}
	writeJSON(w, http.StatusOK, promo)
}

func (h *PromotionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := repository.SoftDeletePromotion(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *PromotionHandler) Active(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	branchIDStr := q.Get("branch_id")

	if branchIDStr != "" {
		bid, err := uuid.Parse(branchIDStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch_id"})
			return
		}
		promos, err := repository.ListPromotionsByBranch(bid)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if promos == nil {
			promos = []model.Promotion{}
		}
		writeJSON(w, http.StatusOK, promos)
		return
	}

	promos, err := repository.GetActivePromotions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if promos == nil {
		promos = []model.Promotion{}
	}
	writeJSON(w, http.StatusOK, promos)
}

func (h *PromotionHandler) ValidateVoucher(w http.ResponseWriter, r *http.Request) {
	var req model.ValidateVoucherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ValidateVoucherResponse{
			Valid: false,
			Error: "invalid request body",
		})
		return
	}
	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, model.ValidateVoucherResponse{
			Valid: false,
			Error: "code is required",
		})
		return
	}

	var branchID *uuid.UUID
	if req.BranchID != "" {
		bid, err := uuid.Parse(req.BranchID)
		if err == nil {
			branchID = &bid
		}
	}

	resp, err := repository.ValidateVoucher(req.Code, branchID, req.Total)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// BranchPromotions returns active promotions for a specific branch.
func (h *PromotionHandler) BranchPromotions(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	promos, err := repository.ListPromotionsByBranch(branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if promos == nil {
		promos = []model.Promotion{}
	}
	writeJSON(w, http.StatusOK, promos)
}
