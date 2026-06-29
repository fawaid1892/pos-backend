package model

import (
	"time"

	"github.com/google/uuid"
)

// User
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"` // owner, admin_cabang, kasir
	BranchID  *uuid.UUID `json:"branch_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type MeResponse struct {
	User User `json:"user"`
}

// Branch
type Branch struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	Phone     string     `json:"phone"`
	TaxRate   float64    `json:"tax_rate"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateBranchRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type UpdateBranchRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// Category
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product
type Product struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Name        string     `json:"name"`
	Barcode     string     `json:"barcode"`
	Price       float64    `json:"price"`
	CostPrice   float64    `json:"cost_price"`
	Stock       int        `json:"stock"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateProductRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Barcode    string    `json:"barcode"`
	Price      float64   `json:"price"`
	CostPrice  float64   `json:"cost_price"`
	Stock      int       `json:"stock"`
}

type UpdateProductRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Barcode    string    `json:"barcode"`
	Price      float64   `json:"price"`
	CostPrice  float64   `json:"cost_price"`
	Stock      int       `json:"stock"`
}

type ProductSearchParams struct {
	Query   string `json:"query"`
	Barcode string `json:"barcode"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

// Transaction
type Transaction struct {
	ID              uuid.UUID  `json:"id"`
	BranchID        uuid.UUID  `json:"branch_id"`
	UserID          uuid.UUID  `json:"user_id"`
	CustomerName    string     `json:"customer_name"`
	Subtotal        float64    `json:"subtotal"`
	DiscountPercent float64    `json:"discount_percent"`
	DiscountAmount  float64    `json:"discount_amount"`
	TaxRate         float64    `json:"tax_rate"`
	TaxAmount       float64    `json:"tax_amount"`
	Total           float64    `json:"total"`
	CashAmount      float64    `json:"cash_amount"`
	ChangeAmount    float64    `json:"change_amount"`
	PaymentMethod    string     `json:"payment_method"`
	PaymentReference string     `json:"payment_reference,omitempty"`
	Items           []TransactionItem `json:"items,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type TransactionItem struct {
	ID          uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Subtotal    float64   `json:"subtotal"`
}

type CheckoutRequest struct {
	BranchID         uuid.UUID         `json:"branch_id"`
	CustomerName     string            `json:"customer_name"`
	DiscountPercent  float64           `json:"discount_percent"`
	PaymentMethod    string            `json:"payment_method"`
	PaymentReference string            `json:"payment_reference,omitempty"`
	CashAmount       float64           `json:"cash_amount"`
	Items            []CheckoutItemReq `json:"items"`
}

type CheckoutItemReq struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// ─── Stock ───

type BranchProduct struct {
	BranchID  uuid.UUID `json:"branch_id"`
	ProductID uuid.UUID `json:"product_id"`
	StockQty  float64   `json:"stock_qty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Joined fields
	ProductName string  `json:"product_name,omitempty"`
	Barcode     string  `json:"barcode,omitempty"`
	Price       float64 `json:"price,omitempty"`
	CostPrice   float64 `json:"cost_price,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

type StockMutation struct {
	ID          uuid.UUID  `json:"id"`
	BranchID    uuid.UUID  `json:"branch_id"`
	ProductID   uuid.UUID  `json:"product_id"`
	Type        string     `json:"type"` // in, out, transfer_in, transfer_out
	Qty         float64    `json:"qty"`
	ReferenceID *uuid.UUID `json:"reference_id,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// Joined fields
	ProductName string `json:"product_name,omitempty"`
	Barcode     string `json:"barcode,omitempty"`
}

type StockAdjustmentRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Type      string    `json:"type"` // in, out
	Qty       float64   `json:"qty"`
	Notes     string    `json:"notes,omitempty"`
}

type StockTransferRequest struct {
	ProductID    uuid.UUID `json:"product_id"`
	SourceBranchID uuid.UUID `json:"source_branch_id"`
	TargetBranchID uuid.UUID `json:"target_branch_id"`
	Qty         float64   `json:"qty"`
	Notes       string    `json:"notes,omitempty"`
}

// ─── Reports ───

type SalesReportRow struct {
	Date        string  `json:"date"`
	TransactionCount int `json:"transaction_count"`
	Subtotal    float64 `json:"subtotal"`
	Discount    float64 `json:"discount"`
	Total       float64 `json:"total"`
}

type StockReportRow struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	Barcode       string    `json:"barcode"`
	CategoryName  string    `json:"category_name"`
	CurrentStock  float64   `json:"current_stock"`
	MinStock      float64   `json:"min_stock,omitempty"`
	LastMutation  *time.Time `json:"last_mutation,omitempty"`
}

// ─── Low Stock ───

type LowStockItem struct {
	ProductName string  `json:"product_name"`
	BranchName  string  `json:"branch_name"`
	StockQty    float64 `json:"stock_qty"`
	MinStock    float64 `json:"min_stock"`
}

// ─── PDF Export ───

type SalesPDFRow struct {
	Date        string  `json:"date"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
	TaxAmount   float64 `json:"tax_amount"`
	Total       float64 `json:"total"`
}

type ProfitLossRow struct {
	ProductID   uuid.UUID  `json:"product_id"`
	ProductName string     `json:"product_name"`
	QtySold     int        `json:"qty_sold"`
	Revenue     float64    `json:"revenue"`
	Cost        float64    `json:"cost"`
	Profit      float64    `json:"profit"`
}

type ProfitLossSummary struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalCost    float64 `json:"total_cost"`
	TotalProfit  float64 `json:"total_profit"`
}

type SalesReportResponse struct {
	Period struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"period"`
	Rows       []SalesReportRow `json:"rows"`
	TotalSales float64          `json:"total_sales"`
	TotalDiscount float64      `json:"total_discount"`
	TotalNet   float64          `json:"total_net"`
	TotalTransactions int       `json:"total_transactions"`
}

type StockReportResponse struct {
	Rows       []StockReportRow `json:"rows"`
	TotalItems int              `json:"total_items"`
}

type ProfitLossReportResponse struct {
	Period struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"period"`
	Rows    []ProfitLossRow  `json:"rows"`
	Summary ProfitLossSummary `json:"summary"`
}

// ─── Export ───

type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatCSV  ExportFormat = "csv"
)

// ─── Dashboard ───

type DashboardStatsResponse struct {
	TodayRevenue      float64 `json:"today_revenue"`
	TotalTransactions int     `json:"total_transactions"`
	ActiveBranches    int     `json:"active_branches"`
	LowStockItems     int     `json:"low_stock_items"`
}

type SalesChartRow struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

type SalesChartResponse struct {
	Period struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"period"`
	Rows []SalesChartRow `json:"rows"`
}

// ─── Generic API Response ───

// APIResponse is a generic JSON response wrapper.
type APIResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message,omitempty"`
	Error   string                  `json:"error,omitempty"`
	Data    interface{}             `json:"data,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// ─── User Management ───

type CreateUserRequest struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	FullName string    `json:"full_name"`
	Role     string    `json:"role"`
	BranchID *uuid.UUID `json:"branch_id,omitempty"`
}

type UpdateUserRequest struct {
	Username  string     `json:"username,omitempty"`
	Password  string     `json:"password,omitempty"`
	FullName  string     `json:"full_name,omitempty"`
	Role      string     `json:"role,omitempty"`
	BranchID  *uuid.UUID `json:"branch_id,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
}

type ListUsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
