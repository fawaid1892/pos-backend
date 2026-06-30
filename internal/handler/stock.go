package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"
	"pos-multi-branch/backend/internal/ws"
)

type StockHandler struct{}

func NewStockHandler() *StockHandler {
	return &StockHandler{}
}

// POST /api/v1/branches/{id}/inventory/adjustment
func (h *StockHandler) Adjustment(w http.ResponseWriter, r *http.Request) {
	branchID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	var req model.StockAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.ProductID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "product_id is required"})
		return
	}
	if req.Type != "in" && req.Type != "out" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type must be 'in' or 'out'"})
		return
	}
	if req.Qty <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "qty must be positive"})
		return
	}

	// Apply adjustment to branch_products
	if req.Type == "in" {
		if err := repository.UpsertBranchProduct(branchID, req.ProductID, req.Qty); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	} else {
		// Check current stock before decreasing
		bp, err := repository.GetBranchProduct(branchID, req.ProductID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if bp == nil || bp.StockQty < req.Qty {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient stock"})
			return
		}
		if err := repository.UpsertBranchProduct(branchID, req.ProductID, -req.Qty); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	// Record stock mutation
	mutation := &model.StockMutation{
		BranchID:  branchID,
		ProductID: req.ProductID,
		Type:      req.Type,
		Qty:       req.Qty,
		Notes:     req.Notes,
	}
	if err := repository.InsertStockMutation(mutation); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Return updated product stock
	bp, err := repository.GetBranchProduct(branchID, req.ProductID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, bp)

	// ── Broadcast stock.adjusted event via WebSocket ──
	if ws.DefaultHub != nil {
		ws.DefaultHub.BroadcastEventToBranch(int64(0), ws.Event{
			Type: ws.EventStockAdjusted,
			Payload: map[string]interface{}{
				"product_id": req.ProductID,
				"type":       req.Type,
				"qty":        req.Qty,
			},
		})
	}

	// Check & broadcast low stock
	go checkAndBroadcastLowStock(branchID)
}

// POST /api/v1/inventory/transfer
func (h *StockHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req model.StockTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.SourceBranchID == 0 || req.TargetBranchID == 0 || req.ProductID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source_branch_id, target_branch_id, and product_id are required"})
		return
	}
	if req.SourceBranchID == req.TargetBranchID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source and target branches must be different"})
		return
	}
	if req.Qty <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "qty must be positive"})
		return
	}

	if err := repository.TransferStock(req.SourceBranchID, req.TargetBranchID, req.ProductID, req.Qty, req.Notes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "transfer successful"})

	// ── Broadcast stock.transferred event via WebSocket ──
	if ws.DefaultHub != nil {
		ws.DefaultHub.BroadcastEvent(ws.Event{
			Type: ws.EventStockTransferred,
			Payload: map[string]interface{}{
				"product_id":       req.ProductID,
				"source_branch_id": req.SourceBranchID,
				"target_branch_id": req.TargetBranchID,
				"qty":              req.Qty,
			},
		})
	}
}

// GET /api/v1/branches/{id}/inventory
func (h *StockHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	branchID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	mutations, err := repository.ListStockMutations(branchID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if mutations == nil {
		mutations = []model.StockMutation{}
	}

	response := map[string]interface{}{
		"mutations": mutations,
	}

	// Also include current stock summary
	products, err := repository.ListBranchProducts(branchID)
	if err == nil {
		response["products"] = products
	}

	writeJSON(w, http.StatusOK, response)
}

// GET /api/v1/branches/{id}/inventory/low-stock?threshold=5
func (h *StockHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	branchID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 5.0
	if thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil && t > 0 {
			threshold = t
		}
	}

	items, err := repository.GetLowStockProducts(branchID, threshold)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if items == nil {
		items = []model.LowStockItem{}
	}

	writeJSON(w, http.StatusOK, items)
}
