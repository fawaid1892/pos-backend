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

<<<<<<< HEAD
	// Database
=======
	// ─── Supabase / PostgreSQL ───
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

<<<<<<< HEAD
=======
	// ─── SQLite sync mirror (local server) ───
	sqliteDB, err := database.NewSQLiteDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("SQLite sync mirror failed: %v", err)
	}
	defer sqliteDB.Close()

>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	// JWT
	middleware.InitJWT(cfg)

	// Handlers
	authH := handler.NewAuthHandler(cfg)
	branchH := handler.NewBranchHandler()
	productH := handler.NewProductHandler()
	txH := handler.NewTransactionHandler()
	stockH := handler.NewStockHandler()
	reportH := handler.NewReportHandler()
	exportH := handler.NewExportHandler()
<<<<<<< HEAD
=======
	syncH := handler.NewSyncHandler(sqliteDB.DB)
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9

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

	// Stock / Inventory
	protected.HandleFunc("POST /api/v1/branches/{id}/inventory/adjustment", stockH.Adjustment)
	protected.HandleFunc("POST /api/v1/inventory/transfer", stockH.Transfer)
	protected.HandleFunc("GET /api/v1/branches/{id}/inventory", stockH.ListInventory)

	// Reports
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/sales", reportH.Sales)
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/stock", reportH.Stock)
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/profit-loss", reportH.ProfitLoss)

	// Export
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/sales/export", exportH.SalesExport)

<<<<<<< HEAD
=======
	// Sync endpoints (authenticated — branches push/pull using their own credentials)
	protected.HandleFunc("POST /api/v1/sync/push", syncH.Push)
	protected.HandleFunc("GET /api/v1/sync/pull", syncH.Pull)
	protected.HandleFunc("POST /api/v1/sync/resolve", syncH.Resolve)

>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	mux.Handle("/api/v1/", middleware.AuthMiddleware(protected))

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
