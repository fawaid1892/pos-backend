package handler

import (
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

	"gorm.io/gorm"
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

	if req.BranchID == 0 {
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
	branch, err := repository.GetBranchByID(req.BranchID)
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
		product, err := repository.GetProductByID(ci.ProductID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if product == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found: " + strconv.FormatInt(ci.ProductID, 10)})
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

	// ── Execute all writes in a single GORM transaction ──
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

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Deduct global stock (products table)
		for _, ci := range req.Items {
			result := tx.Model(&model.Product{}).
				Where("id = ? AND deleted_at IS NULL AND stock >= ?", ci.ProductID, ci.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", ci.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errInsufficientGlobalStock
			}
		}

		// Deduct branch-level stock (branch_products)
		for _, ci := range req.Items {
			result := tx.Model(&model.BranchProduct{}).
				Where("branch_id = ? AND product_id = ? AND stock_qty >= ?", req.BranchID, ci.ProductID, float64(ci.Quantity)).
				UpdateColumn("stock_qty", gorm.Expr("stock_qty - ?", float64(ci.Quantity)))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errInsufficientBranchStock
			}
		}

		// Create transaction
		if err := tx.Create(txData).Error; err != nil {
			return err
		}

		// Insert items
		for i := range items {
			items[i].TransactionID = txData.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		switch err {
		case errInsufficientGlobalStock:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient stock or product not found"})
		case errInsufficientBranchStock:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient stock in this branch or product not found"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	// ── Broadcast transaction.created event via WebSocket ──
	if ws.DefaultHub != nil {
		ws.DefaultHub.BroadcastEvent(ws.Event{
			Type: ws.EventTransactionCreated,
			Payload: map[string]interface{}{
				"transaction_id": txData.ID,
				"branch_id":      txData.BranchID,
				"total":          txData.Total,
				"items_count":    len(items),
				"created_at":     txData.CreatedAt,
			},
		})
	}

	txData.Items = items
	writeJSON(w, http.StatusCreated, txData)

	// ── Check & broadcast low stock via WebSocket ──
	go checkAndBroadcastLowStock(req.BranchID)
}

var (
	errInsufficientGlobalStock = &txError{"insufficient global stock"}
	errInsufficientBranchStock = &txError{"insufficient branch stock"}
)

type txError struct{ msg string }

func (e *txError) Error() string { return e.msg }

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var branchID *int64
	if bidStr := q.Get("branch_id"); bidStr != "" {
		if bid, err := strconv.ParseInt(bidStr, 10, 64); err == nil {
			branchID = &bid
		}
	}

	txs, err := repository.ListTransactions(branchID, limit, offset)
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
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	tx, err := repository.GetTransactionByID(id)
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
func checkAndBroadcastLowStock(branchID int64) {
	lowStock, err := repository.GetLowStockProductsByMinStock(branchID)
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
				"branch_id": branchID,
				"items":     lowStock,
			},
		})
		log.Printf("[stock] low stock alert broadcast for branch %d (%d items)", branchID, len(lowStock))
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
