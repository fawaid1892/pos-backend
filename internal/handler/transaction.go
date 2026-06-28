package handler

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"
	"pos-multi-branch/backend/internal/ws"

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

	// Payment method validation
	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}
	switch paymentMethod {
	case "cash", "qris", "transfer", "edc":
		// valid
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "payment_method must be one of: cash, qris, transfer, edc"})
		return
	}

	if paymentMethod == "cash" && req.CashAmount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cash_amount must be positive for cash payment"})
		return
	}

	userID := middleware.GetUserID(r.Context())

	// ── Fetch branch config for tax_rate ──
	branch, err := repository.GetBranchByID(r.Context(), req.BranchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch branch: " + err.Error()})
		return
	}
	if branch == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "branch not found"})
		return
	}

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
	taxRate := branch.TaxRate
	if taxRate < 0 {
		taxRate = 0
	}

	discountAmount := math.Round(subtotal*discountPercent/100*100) / 100
	afterDiscount := math.Round((subtotal-discountAmount)*100) / 100
	taxAmount := math.Round(afterDiscount*taxRate/100*100) / 100
	total := math.Round((afterDiscount+taxAmount)*100) / 100

	var changeAmount float64
	if paymentMethod == "cash" {
		if req.CashAmount < total {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cash_amount is less than total"})
			return
		}
		changeAmount = math.Round((req.CashAmount-total)*100) / 100
	} else {
		// Non-cash: cash_amount = total, change = 0
		req.CashAmount = total
		changeAmount = 0
	}

	// ── Deduct global stock (products table) ──
	// ── Deduct branch-level stock (branch_products) ──
	// ── Create transaction ──
	// ── Insert items ──
	// Bug D: Wrap all write operations in a single database transaction
	tx, err := database.Pool.Begin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to begin transaction: " + err.Error()})
		return
	}
	defer tx.Rollback(r.Context())

	for _, ci := range req.Items {
		tag, err := tx.Exec(r.Context(),
			`UPDATE products SET stock = stock - $1, updated_at = NOW()
			 WHERE id = $2 AND deleted_at IS NULL AND stock >= $1`,
			ci.Quantity, ci.ProductID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if tag.RowsAffected() == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient stock or product not found"})
			return
		}
	}

	for _, ci := range req.Items {
		tag, err := tx.Exec(r.Context(),
			`UPDATE branch_products SET stock_qty = stock_qty - $1, updated_at = NOW()
			 WHERE branch_id = $2 AND product_id = $3 AND stock_qty >= $1`,
			float64(ci.Quantity), req.BranchID, ci.ProductID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch stock: " + err.Error()})
			return
		}
		if tag.RowsAffected() == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient stock in this branch or product not found"})
			return
		}
	}

	now := r.Context()
	txData := &model.Transaction{
		BranchID:         req.BranchID,
		UserID:           userID,
		CustomerName:     req.CustomerName,
		Subtotal:         subtotal,
		DiscountPercent:  discountPercent,
		DiscountAmount:   discountAmount,
		TaxRate:          taxRate,
		TaxAmount:        taxAmount,
		Total:            total,
		CashAmount:       req.CashAmount,
		ChangeAmount:     changeAmount,
		PaymentMethod:    paymentMethod,
		PaymentReference: req.PaymentReference,
	}

	err = tx.QueryRow(now,
		`INSERT INTO transactions (branch_id, user_id, customer_name, subtotal,
		                           discount_percent, discount_amount, tax_rate, tax_amount, total,
		                           cash_amount, change_amount, payment_method, payment_reference)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING id, created_at`,
		txData.BranchID, txData.UserID, txData.CustomerName,
		txData.Subtotal, txData.DiscountPercent, txData.DiscountAmount,
		txData.TaxRate, txData.TaxAmount,
		txData.Total, txData.CashAmount, txData.ChangeAmount,
		txData.PaymentMethod, txData.PaymentReference,
	).Scan(&txData.ID, &txData.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	for i := range items {
		items[i].TransactionID = txData.ID
		err = tx.QueryRow(now,
			`INSERT INTO transaction_items (transaction_id, product_id, product_name, quantity, price, subtotal)
			 VALUES ($1,$2,$3,$4,$5,$6)
			 RETURNING id`,
			items[i].TransactionID, items[i].ProductID, items[i].ProductName,
			items[i].Quantity, items[i].Price, items[i].Subtotal,
		).Scan(&items[i].ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit transaction: " + err.Error()})
		return
	}

	// ── Broadcast transaction.created event via WebSocket ──
	if ws.DefaultHub != nil {
		ws.DefaultHub.BroadcastEvent(ws.Event{
			Type: ws.EventTransactionCreated,
			Payload: map[string]interface{}{
				"transaction_id": txData.ID.String(),
				"branch_id":      txData.BranchID.String(),
				"total":          txData.Total,
				"items_count":    len(items),
				"created_at":     txData.CreatedAt,
			},
		})
	}

	txData.Items = items
	writeJSON(w, http.StatusCreated, txData)

	// ── Check & broadcast low stock via WebSocket ──
	go checkAndBroadcastLowStock(r.Context(), req.BranchID)
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

// checkAndBroadcastLowStock queries low-stock items for the branch and broadcasts
// a stock.low event via WebSocket if any products are below their min_stock threshold.
func checkAndBroadcastLowStock(ctx context.Context, branchID uuid.UUID) {
	lowStock, err := repository.GetLowStockProductsByMinStock(ctx, branchID)
	if err != nil {
		log.Printf("[stock] failed to check low stock: %v", err)
		return
	}
	if len(lowStock) == 0 {
		return
	}
	if ws.DefaultHub != nil {
		ws.DefaultHub.BroadcastEvent(ws.Event{
			Type: ws.EventStockLow,
			Payload: map[string]interface{}{
				"branch_id": branchID.String(),
				"items":     lowStock,
			},
		})
		log.Printf("[stock] low stock alert broadcast for branch %s (%d items)", branchID.String(), len(lowStock))
	}
}

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
