package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req model.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.BranchID == uuid.Nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one item required"})
		return
	}
	if req.CashAmount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cash_amount must be positive"})
		return
	}

	userID := middleware.GetUserID(r.Context())

	// ── Build transaction items & calculate subtotal ──
	var items []model.TransactionItem
	var subtotal float64

	for _, ci := range req.Items {
		if ci.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid quantity for item"})
			return
		}
		product, err := repository.GetProductByID(r.Context(), ci.ProductID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if product == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found: " + ci.ProductID.String()})
			return
		}

		itemSubtotal := product.Price * float64(ci.Quantity)
		subtotal += itemSubtotal

		items = append(items, model.TransactionItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    ci.Quantity,
			Price:       product.Price,
			Subtotal:    itemSubtotal,
		})
	}

	// ── Calculate discounts & totals ──
	discountPercent := req.DiscountPercent
	if discountPercent < 0 {
		discountPercent = 0
	}
	if discountPercent > 100 {
		discountPercent = 100
	}
	discountAmount := math.Round(subtotal*discountPercent/100*100) / 100
	total := math.Round((subtotal-discountAmount)*100) / 100

	if req.CashAmount < total {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cash_amount is less than total"})
		return
	}
	changeAmount := math.Round((req.CashAmount-total)*100) / 100

	// ── Deduct stock & insert items in a loop (no pgx tx for simplicity) ──
	for _, ci := range req.Items {
		if err := repository.DeductProductStock(r.Context(), ci.ProductID, ci.Quantity); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	// ── Create transaction ──
	now := r.Context()
	tx := &model.Transaction{
		BranchID:        req.BranchID,
		UserID:          userID,
		CustomerName:    req.CustomerName,
		Subtotal:        subtotal,
		DiscountPercent: discountPercent,
		DiscountAmount:  discountAmount,
		Total:           total,
		CashAmount:      req.CashAmount,
		ChangeAmount:    changeAmount,
	}

	if err := repository.CreateTransaction(now, tx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// ── Insert items ──
	for i := range items {
		items[i].TransactionID = tx.ID
		if err := repository.InsertTransactionItem(now, &items[i]); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	tx.Items = items
	writeJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var branchID *uuid.UUID
	if bidStr := q.Get("branch_id"); bidStr != "" {
		if bid, err := uuid.Parse(bidStr); err == nil {
			branchID = &bid
		}
	}

	txs, err := repository.ListTransactions(r.Context(), branchID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if txs == nil {
		txs = []model.Transaction{}
	}
	writeJSON(w, http.StatusOK, txs)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	tx, err := repository.GetTransactionByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if tx == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// ─── JSON Helper ───

// writeJSON writes a JSON response safely, preventing panics from encoder failures.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Response already sent (or partially sent) — log and move on.
		// This cannot panic; Encode returns error codes for type issues.
		_ = err
	}
}
