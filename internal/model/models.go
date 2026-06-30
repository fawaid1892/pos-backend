package model

import (
	"time"

	"gorm.io/gorm"
)

// ─── User ───
type User struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null;size:100"`
	Password  string         `json:"-" gorm:"not null"`
	FullName  string         `json:"full_name" gorm:"size:200;default:''"`
	Role      string         `json:"role" gorm:"size:20;default:kasir"` // owner, admin_cabang, kasir
	RoleID    *int64         `json:"role_id,omitempty"`
	BranchID  *int64         `json:"branch_id,omitempty"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         User   `json:"user"`
}

type RefreshToken struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	UserID    int64      `json:"user_id" gorm:"not null;index"`
	Token     string     `json:"token" gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type MeResponse struct {
	User User `json:"user"`
}

// ─── Branch ───
type Branch struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null;size:200"`
	Code         string         `json:"code" gorm:"uniqueIndex;size:50;not null;default:''"`
	Address      string         `json:"address" gorm:"size:500;default:''"`
	Phone        string         `json:"phone" gorm:"size:30;default:''"`
	Province     string         `json:"province" gorm:"size:100;not null;default:''"`
	ProvinceCode string         `json:"province_code" gorm:"size:10;not null;default:''"`
	City         string         `json:"city" gorm:"size:100;not null;default:''"`
	CityCode     string         `json:"city_code" gorm:"size:10;not null;default:''"`
	TaxRate      float64        `json:"tax_rate" gorm:"default:0"`
	IsActive     bool           `json:"is_active" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type CreateBranchRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Province     string `json:"province"`
	ProvinceCode string `json:"province_code"`
	City         string `json:"city"`
	CityCode     string `json:"city_code"`
}

type UpdateBranchRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Province     string `json:"province"`
	ProvinceCode string `json:"province_code"`
	City         string `json:"city"`
	CityCode     string `json:"city_code"`
}

// ─── Category ───
type Category struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null;size:100"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ─── Product ───
type Product struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	CategoryID   int64          `json:"category_id" gorm:"not null"`
	CategoryName string         `json:"category_name,omitempty" gorm:"-:all"` // joined field, not stored
	Name         string         `json:"name" gorm:"not null;size:200;index"`
	Code         *string        `json:"code" gorm:"uniqueIndex;size:100;default:null"`
	Barcode      *string        `json:"barcode" gorm:"uniqueIndex;size:100;default:null"`
	Unit         string         `json:"unit" gorm:"size:20;default:PCS"`
	Price        float64        `json:"price" gorm:"not null;default:0"`
	CostPrice    float64        `json:"cost_price" gorm:"default:0"`
	Stock        int            `json:"stock" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type CreateProductRequest struct {
	CategoryID int64   `json:"category_id"`
	Name       string  `json:"name"`
	Code       *string `json:"code,omitempty"`
	Barcode    *string `json:"barcode,omitempty"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	CostPrice  float64 `json:"cost_price"`
	Stock      int     `json:"stock"`
}

type UpdateProductRequest struct {
	CategoryID int64   `json:"category_id"`
	Name       string  `json:"name"`
	Code       *string `json:"code,omitempty"`
	Barcode    *string `json:"barcode,omitempty"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	CostPrice  float64 `json:"cost_price"`
	Stock      int     `json:"stock"`
}

type ProductSearchParams struct {
	Query   string `json:"query"`
	Barcode string `json:"barcode"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

// ─── Transaction ───
type Transaction struct {
	ID               int64             `json:"id" gorm:"primaryKey"`
	BranchID         int64             `json:"branch_id" gorm:"not null;index"`
	UserID           int64             `json:"user_id" gorm:"not null"`
	CustomerName     string            `json:"customer_name" gorm:"size:200;default:''"`
	Subtotal         float64           `json:"subtotal" gorm:"not null;default:0"`
	DiscountPercent  float64           `json:"discount_percent" gorm:"default:0"`
	DiscountAmount   float64           `json:"discount_amount" gorm:"default:0"`
	TaxRate          float64           `json:"tax_rate" gorm:"default:0"`
	TaxAmount        float64           `json:"tax_amount" gorm:"default:0"`
	Total            float64           `json:"total" gorm:"not null;default:0"`
	CashAmount       float64           `json:"cash_amount" gorm:"default:0"`
	ChangeAmount     float64           `json:"change_amount" gorm:"default:0"`
	PaymentMethod    string            `json:"payment_method" gorm:"size:20;default:cash"`
	PaymentReference string            `json:"payment_reference,omitempty" gorm:"size:200;default:''"`
	Items            []TransactionItem `json:"items,omitempty" gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	CreatedAt        time.Time         `json:"created_at" gorm:"autoCreateTime;index"`
}

type TransactionItem struct {
	ID            int64   `json:"id" gorm:"primaryKey"`
	TransactionID int64   `json:"transaction_id" gorm:"not null;index"`
	ProductID     int64   `json:"product_id" gorm:"not null"`
	ProductName   string  `json:"product_name" gorm:"size:200;default:''"`
	Quantity      int     `json:"quantity" gorm:"not null;default:0"`
	Price         float64 `json:"price" gorm:"not null;default:0"`
	Subtotal      float64 `json:"subtotal" gorm:"not null;default:0"`
}

type CheckoutRequest struct {
	BranchID         int64             `json:"branch_id"`
	CustomerName     string            `json:"customer_name"`
	DiscountPercent  float64           `json:"discount_percent"`
	PaymentMethod    string            `json:"payment_method"`
	PaymentReference string            `json:"payment_reference,omitempty"`
	CashAmount       float64           `json:"cash_amount"`
	Items            []CheckoutItemReq `json:"items"`
}

type CheckoutItemReq struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

// ─── Stock ───

type BranchProduct struct {
	BranchID  int64     `json:"branch_id" gorm:"primaryKey"`
	ProductID int64     `json:"product_id" gorm:"primaryKey"`
	StockQty  float64   `json:"stock_qty" gorm:"default:0"`
	MinStock  float64   `json:"min_stock" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Joined fields (not stored in DB)
	ProductName  string  `json:"product_name,omitempty" gorm:"-:all"`
	Barcode      string  `json:"barcode,omitempty" gorm:"-:all"`
	Price        float64 `json:"price,omitempty" gorm:"-:all"`
	CostPrice    float64 `json:"cost_price,omitempty" gorm:"-:all"`
	CategoryName string  `json:"category_name,omitempty" gorm:"-:all"`
}

type StockMutation struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	BranchID    int64      `json:"branch_id" gorm:"not null;index"`
	ProductID   int64      `json:"product_id" gorm:"not null"`
	Type        string     `json:"type" gorm:"size:20;not null"` // in, out, transfer_in, transfer_out
	Qty         float64    `json:"qty" gorm:"not null;default:0"`
	ReferenceID *int64     `json:"reference_id,omitempty"`
	Notes       string     `json:"notes,omitempty" gorm:"size:500;default:''"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Joined fields
	ProductName string `json:"product_name,omitempty" gorm:"-:all"`
	Barcode     string `json:"barcode,omitempty" gorm:"-:all"`
}

type StockAdjustmentRequest struct {
	ProductID int64   `json:"product_id"`
	Type      string  `json:"type"` // in, out
	Qty       float64 `json:"qty"`
	Notes     string  `json:"notes,omitempty"`
}

type StockTransferRequest struct {
	ProductID      int64   `json:"product_id"`
	SourceBranchID int64   `json:"source_branch_id"`
	TargetBranchID int64   `json:"target_branch_id"`
	Qty            float64 `json:"qty"`
	Notes          string  `json:"notes,omitempty"`
}

// ─── Reports ───

type SalesReportRow struct {
	Date              string  `json:"date"`
	TransactionCount  int     `json:"transaction_count"`
	Subtotal          float64 `json:"subtotal"`
	Discount          float64 `json:"discount"`
	Total             float64 `json:"total"`
}

type StockReportRow struct {
	ProductID     int64      `json:"product_id"`
	ProductName   string     `json:"product_name"`
	Barcode       string     `json:"barcode"`
	CategoryName  string     `json:"category_name"`
	CurrentStock  float64    `json:"current_stock"`
	MinStock      float64    `json:"min_stock,omitempty"`
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
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtySold     int     `json:"qty_sold"`
	Revenue     float64 `json:"revenue"`
	Cost        float64 `json:"cost"`
	Profit      float64 `json:"profit"`
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
	Rows              []SalesReportRow `json:"rows"`
	TotalSales        float64          `json:"total_sales"`
	TotalDiscount     float64          `json:"total_discount"`
	TotalNet          float64          `json:"total_net"`
	TotalTransactions int              `json:"total_transactions"`
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

// ─── RBAC ───

type Permission struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Label     string    `json:"label" gorm:"size:200;default:''"`
	Group     string    `json:"group" gorm:"size:50;default:''"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Role struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null;size:50"`
	Description string         `json:"description" gorm:"size:255;default:''"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type RolePermission struct {
	RoleID       int64 `json:"role_id" gorm:"primaryKey"`
	PermissionID int64 `json:"permission_id" gorm:"primaryKey"`
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
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	BranchID *int64 `json:"branch_id,omitempty"`
}

type UpdateUserRequest struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	Role      string `json:"role,omitempty"`
	BranchID  *int64 `json:"branch_id,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

type ListUsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// ─── PromotionBranch (join table) ───

type PromotionBranch struct {
	PromotionID int64 `json:"promotion_id" gorm:"primaryKey;not null"`
	BranchID    int64 `json:"branch_id" gorm:"primaryKey;not null"`
}

// ─── Promotion ───

type Promotion struct {
	ID            int64          `json:"id" gorm:"primaryKey"`
	Name          string         `json:"name" gorm:"not null;size:200"`
	Type          string         `json:"type" gorm:"not null;size:20"`
	Code          *string        `json:"code,omitempty" gorm:"uniqueIndex;size:50;default:null"`
	DiscountValue float64        `json:"discount_value" gorm:"not null;default:0"`
	DiscountType  string         `json:"discount_type" gorm:"not null;size:10;default:persen"`
	SkuTarget     *string        `json:"sku_target,omitempty" gorm:"size:100;default:null"`
	QtyMin        int            `json:"qty_min" gorm:"default:0"`
	QtyFree       int            `json:"qty_free" gorm:"default:0"`
	StartDate     time.Time      `json:"start_date" gorm:"not null"`
	EndDate       time.Time      `json:"end_date" gorm:"not null"`
	Scope         string         `json:"scope" gorm:"size:20;default:selected"` // all, province, city, selected
	ProvinceID    string         `json:"province_id,omitempty" gorm:"size:10;default:''"` // untuk scope=province/city
	CityID        string         `json:"city_id,omitempty" gorm:"size:10;default:''"`     // untuk scope=city
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	MaxUses       int            `json:"max_uses" gorm:"default:0"`
	CurrentUses   int            `json:"current_uses" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	// Relasi
	Branches      []PromotionBranch `json:"branches,omitempty" gorm:"foreignKey:PromotionID"`
}

type CreatePromotionRequest struct {
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Code          *string   `json:"code,omitempty"`
	DiscountValue float64   `json:"discount_value"`
	DiscountType  string    `json:"discount_type"`
	SkuTarget     *string   `json:"sku_target,omitempty"`
	QtyMin        int       `json:"qty_min"`
	QtyFree       int       `json:"qty_free"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Scope         string    `json:"scope"`       // all, province, city, selected
	ProvinceID    string    `json:"province_id,omitempty"`
	CityID        string    `json:"city_id,omitempty"`
	BranchIDs     []int64   `json:"branch_ids,omitempty"` // int64 IDs untuk scope=selected
	IsActive      *bool     `json:"is_active,omitempty"`
	MaxUses       int       `json:"max_uses"`
}

type UpdatePromotionRequest struct {
	Name          *string    `json:"name,omitempty"`
	Type          *string    `json:"type,omitempty"`
	Code          *string    `json:"code,omitempty"`
	DiscountValue *float64   `json:"discount_value,omitempty"`
	DiscountType  *string    `json:"discount_type,omitempty"`
	SkuTarget     *string    `json:"sku_target,omitempty"`
	QtyMin        *int       `json:"qty_min,omitempty"`
	QtyFree       *int       `json:"qty_free,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Scope         *string    `json:"scope,omitempty"`
	ProvinceID    *string    `json:"province_id,omitempty"`
	CityID        *string    `json:"city_id,omitempty"`
	BranchIDs     []int64    `json:"branch_ids,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	MaxUses       *int       `json:"max_uses,omitempty"`
}

type ValidateVoucherRequest struct {
	Code     string  `json:"code"`
	BranchID string  `json:"branch_id"`
	Total    float64 `json:"total"`
}

type ValidateVoucherResponse struct {
	Valid         bool    `json:"valid"`
	DiscountValue float64 `json:"discount_value,omitempty"`
	DiscountType  string  `json:"discount_type,omitempty"`
	PromotionName string  `json:"promotion_name,omitempty"`
	Error         string  `json:"error,omitempty"`
}
