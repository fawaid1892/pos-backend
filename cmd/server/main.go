package main

import (
	"log"
	"net/http"

	"pos-multi-branch/backend/internal/config"
	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/handler"
	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/ws"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "pos-multi-branch/backend/docs"
)

// @title           POS Multi-Branch API
// @version         1.0
// @description     Point of Sale multi-branch backend API with inventory management, reporting, and sync capabilities.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@sukrawetan.ai

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".
func main() {
	cfg := config.Load()

	// ─── Supabase / PostgreSQL ───
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// ─── SQLite sync mirror (local server) ───
	sqliteDB, err := database.NewSQLiteDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("SQLite sync mirror failed: %v", err)
	}
	defer sqliteDB.Close()

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
	userH := handler.NewUserHandler()
	syncH := handler.NewSyncHandler(sqliteDB.DB)
	dashboardH := handler.NewDashboardHandler()

	// WebSocket hub for realtime notifications
	wsHub := ws.NewHub()
	ws.SetDefaultHub(wsHub)

	mux := http.NewServeMux()

	// ─── Swagger UI ───
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// ─── Public routes ───
	mux.Handle("GET /api/v1/ws", wsHub)
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
	protected.HandleFunc("GET /api/v1/branches/{id}/inventory/low-stock", stockH.LowStock)

	// Reports
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/sales", reportH.Sales)
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/sales.pdf", reportH.SalesPDF)
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/stock", reportH.Stock)
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/profit-loss", reportH.ProfitLoss)

	// Export
	protected.HandleFunc("GET /api/v1/branches/{id}/reports/sales/export", exportH.SalesExport)

	// Users
	protected.HandleFunc("GET /api/v1/users", userH.List)
	protected.HandleFunc("GET /api/v1/users/{id}", userH.GetByID)
	protected.HandleFunc("POST /api/v1/users", userH.Create)
	protected.HandleFunc("PUT /api/v1/users/{id}", userH.Update)
	protected.HandleFunc("DELETE /api/v1/users/{id}", userH.Delete)

	// Sync endpoints (authenticated — branches push/pull using their own credentials)
	protected.HandleFunc("POST /api/v1/sync/push", syncH.Push)
	protected.HandleFunc("GET /api/v1/sync/pull", syncH.Pull)
	protected.HandleFunc("POST /api/v1/sync/resolve", syncH.Resolve)

	// Dashboard
	protected.HandleFunc("GET /api/v1/dashboard/stats", dashboardH.DashboardStats)
	protected.HandleFunc("GET /api/v1/dashboard/sales-chart", dashboardH.SalesChart)

	mux.Handle("/api/v1/", middleware.AuthMiddleware(protected))

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
