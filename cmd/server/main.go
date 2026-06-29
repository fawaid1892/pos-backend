package main

import (
	"log"
	"net/http"

	"pos-multi-branch/backend/internal/config"
	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/handler"
	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/ws"
)

func main() {
	cfg := config.Load()

	// ─── PostgreSQL (ElectricSQL-managed via logical replication) ───
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// ─── AutoMigrate ───
	if err := database.Migrate(); err != nil {
		log.Printf("[warn] Database migration: %v (tables may already exist)", err)
	}

	// ElectricSQL — shapes managed via dashboard not API
	log.Printf("ElectricSQL URL: %s (shapes managed via dashboard)", cfg.ElectricURL)

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
	dashboardH := handler.NewDashboardHandler()
	electricH := handler.NewElectricHandler()

	// WebSocket hub for realtime notifications
	wsHub := ws.NewHub()
	ws.SetDefaultHub(wsHub)

	mux := http.NewServeMux()

	// ─── Public routes ───
	public := http.NewServeMux()
	public.Handle("GET /api/v1/ws", wsHub)
	public.HandleFunc("POST /api/v1/auth/login", authH.Login)
	public.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	// ─── Protected routes (with auth middleware) ───
	protected := http.NewServeMux()

	// RBAC wrappers for role-based access control
	adminOnly := middleware.RequireRole("admin_cabang")
	adminOrOwner := middleware.RequireRole("admin_cabang", "owner")
	kasirOrAdmin := middleware.RequireRole("kasir", "admin_cabang")

	// Auth
	protected.HandleFunc("GET /api/v1/auth/me", authH.Me)
	protected.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// Branches CRUD → admin only
	protected.Handle("GET /api/v1/branches", adminOnly(http.HandlerFunc(branchH.List)))
	protected.Handle("GET /api/v1/branches/{id}", adminOnly(http.HandlerFunc(branchH.GetByID)))
	protected.Handle("POST /api/v1/branches", adminOnly(http.HandlerFunc(branchH.Create)))
	protected.Handle("PUT /api/v1/branches/{id}", adminOnly(http.HandlerFunc(branchH.Update)))
	protected.Handle("DELETE /api/v1/branches/{id}", adminOnly(http.HandlerFunc(branchH.Delete)))

	// Products → list & detail for kasir+admin, create/update/delete for admin only
	protected.Handle("GET /api/v1/products", kasirOrAdmin(http.HandlerFunc(productH.List)))
	protected.Handle("GET /api/v1/products/{id}", kasirOrAdmin(http.HandlerFunc(productH.GetByID)))
	protected.Handle("POST /api/v1/products", adminOnly(http.HandlerFunc(productH.Create)))
	protected.Handle("PUT /api/v1/products/{id}", adminOnly(http.HandlerFunc(productH.Update)))
	protected.Handle("DELETE /api/v1/products/{id}", adminOnly(http.HandlerFunc(productH.Delete)))

	// Categories → list for kasir+admin, create for admin only
	protected.Handle("GET /api/v1/categories", kasirOrAdmin(http.HandlerFunc(productH.ListCategories)))
	protected.Handle("POST /api/v1/categories", adminOnly(http.HandlerFunc(productH.CreateCategory)))

	// Transactions → kasir + admin
	protected.Handle("GET /api/v1/transactions", kasirOrAdmin(http.HandlerFunc(txH.List)))
	protected.Handle("GET /api/v1/transactions/{id}", kasirOrAdmin(http.HandlerFunc(txH.GetByID)))
	protected.Handle("POST /api/v1/transactions/checkout", kasirOrAdmin(http.HandlerFunc(txH.Checkout)))

	// Stock / Inventory → admin only (manage stock)
	protected.Handle("POST /api/v1/branches/{id}/inventory/adjustment", adminOnly(http.HandlerFunc(stockH.Adjustment)))
	protected.Handle("POST /api/v1/inventory/transfer", adminOnly(http.HandlerFunc(stockH.Transfer)))
	protected.Handle("GET /api/v1/branches/{id}/inventory", adminOnly(http.HandlerFunc(stockH.ListInventory)))
	protected.Handle("GET /api/v1/branches/{id}/inventory/low-stock", adminOnly(http.HandlerFunc(stockH.LowStock)))

	// Reports → admin + owner
	protected.Handle("GET /api/v1/branches/{id}/reports/sales", adminOrOwner(http.HandlerFunc(reportH.Sales)))
	protected.Handle("GET /api/v1/branches/{id}/reports/sales.pdf", adminOrOwner(http.HandlerFunc(reportH.SalesPDF)))
	protected.Handle("GET /api/v1/branches/{id}/reports/stock", adminOrOwner(http.HandlerFunc(reportH.Stock)))
	protected.Handle("GET /api/v1/branches/{id}/reports/profit-loss", adminOrOwner(http.HandlerFunc(reportH.ProfitLoss)))

	// Export → admin + owner
	protected.Handle("GET /api/v1/branches/{id}/reports/sales/export", adminOrOwner(http.HandlerFunc(exportH.SalesExport)))

	// Users management → admin only
	protected.Handle("GET /api/v1/users", adminOnly(http.HandlerFunc(userH.List)))
	protected.Handle("GET /api/v1/users/{id}", adminOnly(http.HandlerFunc(userH.GetByID)))
	protected.Handle("POST /api/v1/users", adminOnly(http.HandlerFunc(userH.Create)))
	protected.Handle("PUT /api/v1/users/{id}", adminOnly(http.HandlerFunc(userH.Update)))
	protected.Handle("DELETE /api/v1/users/{id}", adminOnly(http.HandlerFunc(userH.Delete)))

	// Dashboard → admin + owner only
	protected.Handle("GET /api/v1/dashboard/stats", adminOrOwner(http.HandlerFunc(dashboardH.DashboardStats)))
	protected.Handle("GET /api/v1/dashboard/sales-chart", adminOrOwner(http.HandlerFunc(dashboardH.SalesChart)))

	// ElectricSQL shape status
	protected.HandleFunc("GET /api/v1/electric/shapes", electricH.Shapes)

	// ─── Top-level dispatcher: public routes first, then auth-protected ───
	mux.Handle("/api/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public routes — handle directly without auth
		if r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/auth/refresh" ||
			(r.URL.Path == "/api/v1/ws" && r.Method == "GET") {
			public.ServeHTTP(w, r)
			return
		}
		// Everything else under /api/v1/ requires auth
		middleware.AuthMiddleware(protected).ServeHTTP(w, r)
	}))

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
