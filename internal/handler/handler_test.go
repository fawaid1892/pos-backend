package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos-multi-branch/backend/internal/model"

	"github.com/google/uuid"
)

// ============================================================
// TRANSACTION HANDLER TESTS
// ============================================================

func TestCheckout_Validation_EmptyBody(t *testing.T) {
	h := NewTransactionHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/checkout", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Checkout(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Checkout(empty body) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCheckout_Validation_MissingBranchID(t *testing.T) {
	h := NewTransactionHandler()
	body, _ := json.Marshal(model.CheckoutRequest{
		BranchID:        uuid.Nil,
		CustomerName:    "Test",
		DiscountPercent: 0,
		CashAmount:      50000,
		Items:           []model.CheckoutItemReq{{ProductID: uuid.New(), Quantity: 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Checkout(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Checkout(no branch_id) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCheckout_Validation_NoItems(t *testing.T) {
	h := NewTransactionHandler()
	body, _ := json.Marshal(model.CheckoutRequest{
		BranchID:        uuid.New(),
		CustomerName:    "Test",
		DiscountPercent: 0,
		CashAmount:      50000,
		Items:           []model.CheckoutItemReq{},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Checkout(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Checkout(no items) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCheckout_Validation_CashAmountZero(t *testing.T) {
	h := NewTransactionHandler()
	body, _ := json.Marshal(model.CheckoutRequest{
		BranchID:        uuid.New(),
		CustomerName:    "Test",
		DiscountPercent: 0,
		CashAmount:      0,
		Items:           []model.CheckoutItemReq{{ProductID: uuid.New(), Quantity: 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Checkout(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Checkout(cash=0) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCheckout_Validation_InvalidItemQuantity(t *testing.T) {
	// NOTE: This test is skipped because item-quantity validation happens AFTER
	// repository.GetBranchByID (line 63 of transaction.go), which requires a
	// live database connection. Without database.Pool, the handler panics.
	t.Skip("Requires DB connection — item quantity validation runs after branch fetch")
	h := NewTransactionHandler()
	tests := []struct {
		name  string
		items []model.CheckoutItemReq
	}{
		{"zero qty", []model.CheckoutItemReq{{ProductID: uuid.New(), Quantity: 0}}},
		{"negative qty", []model.CheckoutItemReq{{ProductID: uuid.New(), Quantity: -1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(model.CheckoutRequest{
				BranchID:        uuid.New(),
				CustomerName:    "Test",
				DiscountPercent: 0,
				CashAmount:      50000,
				Items:           tt.items,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/checkout", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.Checkout(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("Checkout(%s) = %d, want 400; body=%s", tt.name, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGetTransactionByID_InvalidUUID(t *testing.T) {
	h := NewTransactionHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/{id}", nil)
	req.SetPathValue("id", "not-a-uuid")
	rr := httptest.NewRecorder()
	h.GetByID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetByID(invalid) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

// ============================================================
// STOCK HANDLER TESTS
// ============================================================

func TestAdjustment_InvalidBranchID(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockAdjustmentRequest{
		ProductID: uuid.New(),
		Type:      "in",
		Qty:       10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches/{id}/inventory/adjustment", bytes.NewReader(body))
	req.SetPathValue("id", "bad-uuid")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Adjustment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Adjustment(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdjustment_InvalidType(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockAdjustmentRequest{
		ProductID: uuid.New(),
		Type:      "badtype",
		Qty:       10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches/{id}/inventory/adjustment", bytes.NewReader(body))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Adjustment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Adjustment(bad type) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdjustment_ZeroQty(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockAdjustmentRequest{
		ProductID: uuid.New(),
		Type:      "in",
		Qty:       0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches/{id}/inventory/adjustment", bytes.NewReader(body))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Adjustment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Adjustment(zero qty) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdjustment_NegativeQty(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockAdjustmentRequest{
		ProductID: uuid.New(),
		Type:      "in",
		Qty:       -5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches/{id}/inventory/adjustment", bytes.NewReader(body))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Adjustment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Adjustment(neg qty) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdjustment_MissingProductID(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockAdjustmentRequest{
		ProductID: uuid.Nil,
		Type:      "in",
		Qty:       10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches/{id}/inventory/adjustment", bytes.NewReader(body))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Adjustment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Adjustment(missing product) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestTransfer_SameBranch(t *testing.T) {
	h := NewStockHandler()
	id := uuid.New()
	body, _ := json.Marshal(model.StockTransferRequest{
		SourceBranchID: id,
		TargetBranchID: id,
		ProductID:      uuid.New(),
		Qty:            5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Transfer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Transfer(same branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestTransfer_ZeroQty(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockTransferRequest{
		SourceBranchID: uuid.New(),
		TargetBranchID: uuid.New(),
		ProductID:      uuid.New(),
		Qty:            0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Transfer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Transfer(zero qty) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestTransfer_MissingSource(t *testing.T) {
	h := NewStockHandler()
	body, _ := json.Marshal(model.StockTransferRequest{
		SourceBranchID: uuid.Nil,
		TargetBranchID: uuid.New(),
		ProductID:      uuid.New(),
		Qty:            5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Transfer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Transfer(missing source) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestListInventory_InvalidBranchID(t *testing.T) {
	h := NewStockHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/inventory", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.ListInventory(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ListInventory(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

// ============================================================
// REPORT HANDLER TESTS
// ============================================================

func TestSalesReport_InvalidBranchID(t *testing.T) {
	h := NewReportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/sales", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.Sales(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Sales report(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestSalesReport_InvalidStartDate(t *testing.T) {
	h := NewReportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/sales?start=notadate", nil)
	req.SetPathValue("id", uuid.New().String())
	rr := httptest.NewRecorder()
	h.Sales(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Sales report(bad date) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestStockReport_InvalidBranchID(t *testing.T) {
	h := NewReportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/stock", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.Stock(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Stock report(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestProfitLoss_InvalidBranchID(t *testing.T) {
	h := NewReportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/profit-loss", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.ProfitLoss(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Profit-loss(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestProfitLoss_InvalidDate(t *testing.T) {
	h := NewReportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/profit-loss?start=bad", nil)
	req.SetPathValue("id", uuid.New().String())
	rr := httptest.NewRecorder()
	h.ProfitLoss(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Profit-loss(bad date) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

// ============================================================
// PRODUCT HANDLER TESTS
// ============================================================

func TestCreateProduct_EmptyBody(t *testing.T) {
	h := NewProductHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateProduct(empty) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateProduct_MissingName(t *testing.T) {
	h := NewProductHandler()
	body, _ := json.Marshal(model.CreateProductRequest{
		CategoryID: uuid.New(),
		Name:       "",
		Barcode:    "12345",
		Price:      10000,
		Stock:      10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateProduct(no name) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateProduct_MissingBarcode(t *testing.T) {
	h := NewProductHandler()
	body, _ := json.Marshal(model.CreateProductRequest{
		CategoryID: uuid.New(),
		Name:       "Test",
		Barcode:    "",
		Price:      10000,
		Stock:      10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateProduct(no barcode) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetProductByID_InvalidUUID(t *testing.T) {
	h := NewProductHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/{id}", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.GetByID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetByID(bad id) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestUpdateProduct_InvalidUUID(t *testing.T) {
	h := NewProductHandler()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/{id}", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.Update(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Update(bad id) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteProduct_InvalidUUID(t *testing.T) {
	h := NewProductHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/{id}", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Delete(bad id) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateCategory_EmptyBody(t *testing.T) {
	h := NewProductHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateCategory(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateCategory(empty) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateCategory_EmptyName(t *testing.T) {
	h := NewProductHandler()
	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateCategory(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("CreateCategory(no name) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

// ============================================================
// EXPORT HANDLER TESTS
// ============================================================

func TestSalesExport_InvalidBranchID(t *testing.T) {
	h := NewExportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/sales/export", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()
	h.SalesExport(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Export(bad branch) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestSalesExport_InvalidDate(t *testing.T) {
	h := NewExportHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/{id}/reports/sales/export?start=bad", nil)
	req.SetPathValue("id", uuid.New().String())
	rr := httptest.NewRecorder()
	h.SalesExport(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Export(bad date) = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestSalesExport_UnsupportedFormat(t *testing.T) {
	// Note: This test requires a DB connection because format validation
	// happens after DB fetch. Skipping until integration test setup.
	t.Skip("Requires DB connection — format validation runs after DB call")
}
