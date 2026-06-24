package main

import (
	"log"
	"net/http"

	"pos-multi-branch/backend/internal/config"
	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/handler"
	"pos-multi-branch/backend/internal/middleware"
)

func main() {
	cfg := config.Load()

	// Database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// JWT
	middleware.InitJWT(cfg)

	// Handlers
	authH := handler.NewAuthHandler(cfg)
	branchH := handler.NewBranchHandler()
	productH := handler.NewProductHandler()
	txH := handler.NewTransactionHandler()

	mux := http.NewServeMux()

	// ─── Public routes ───
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	// ─── Protected routes ───
	protected := http.NewServeMux()

	// Auth
	protected.HandleFunc("GET /api/v1/auth/me", authH.Me)
	protected.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// Branches
	protected.HandleFunc("GET /api/v1/branches", branchH.List)
	protected.HandleFunc("GET /api/v1/branches/{id}", branchH.GetByID)
	protected.HandleFunc("POST /api/v1/branches", branchH.Create)
	protected.HandleFunc("PUT /api/v1/branches/{id}", branchH.Update)
	protected.HandleFunc("DELETE /api/v1/branches/{id}", branchH.Delete)

	// Products
	protected.HandleFunc("GET /api/v1/products", productH.List)
	protected.HandleFunc("GET /api/v1/products/{id}", productH.GetByID)
	protected.HandleFunc("POST /api/v1/products", productH.Create)
	protected.HandleFunc("PUT /api/v1/products/{id}", productH.Update)
	protected.HandleFunc("DELETE /api/v1/products/{id}", productH.Delete)

	// Categories
	protected.HandleFunc("GET /api/v1/categories", productH.ListCategories)
	protected.HandleFunc("POST /api/v1/categories", productH.CreateCategory)

	// Transactions
	protected.HandleFunc("GET /api/v1/transactions", txH.List)
	protected.HandleFunc("GET /api/v1/transactions/{id}", txH.GetByID)
	protected.HandleFunc("POST /api/v1/transactions/checkout", txH.Checkout)

	mux.Handle("/api/v1/", middleware.AuthMiddleware(protected))

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
