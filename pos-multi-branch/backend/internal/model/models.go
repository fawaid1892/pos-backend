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
	Role      string    `json:"role"` // admin, kasir, owner
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
	Stock      int       `json:"stock"`
}

type UpdateProductRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Barcode    string    `json:"barcode"`
	Price      float64   `json:"price"`
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
	Total           float64    `json:"total"`
	CashAmount      float64    `json:"cash_amount"`
	ChangeAmount    float64    `json:"change_amount"`
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
	BranchID        uuid.UUID         `json:"branch_id"`
	CustomerName    string            `json:"customer_name"`
	DiscountPercent float64           `json:"discount_percent"`
	CashAmount      float64           `json:"cash_amount"`
	Items           []CheckoutItemReq `json:"items"`
}

type CheckoutItemReq struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
