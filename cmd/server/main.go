package main

import (
	"log"
	"net/http"

	"pos-multi-branch/backend/internal/config"
	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/handler"
	"pos-multi-branch/backend/internal/middleware"
	"pos-multi-branch/backend/internal/repository"
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
	roleH := handler.NewRoleHandler()

	// Wire up the RoleHasPermission function (avoids circular import)
	middleware.RoleHasPermission = repository.RoleHasPermission

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

	// requirePerm is a shorthand for permission-based RBAC middleware
	requirePerm := middleware.RequirePermission

	// Auth
	protected.HandleFunc("GET /api/v1/auth/me", authH.Me)
	protected.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// Branches CRUD → branches.*
	protected.Handle("GET /api/v1/branches", requirePerm("branches.read")(http.HandlerFunc(branchH.List)))
	protected.Handle("GET /api/v1/branches/{id}", requirePerm("branches.read")(http.HandlerFunc(branchH.GetByID)))
	protected.Handle("POST /api/v1/branches", requirePerm("branches.create")(http.HandlerFunc(branchH.Create)))
	protected.Handle("PUT /api/v1/branches/{id}", requirePerm("branches.update")(http.HandlerFunc(branchH.Update)))
	protected.Handle("DELETE /api/v1/branches/{id}", requirePerm("branches.delete")(http.HandlerFunc(branchH.Delete)))

	// Products → products.*
	protected.Handle("GET /api/v1/products", requirePerm("products.read")(http.HandlerFunc(productH.List)))
	protected.Handle("GET /api/v1/products/{id}", requirePerm("products.read")(http.HandlerFunc(productH.GetByID)))
	protected.Handle("POST /api/v1/products", requirePerm("products.create")(http.HandlerFunc(productH.Create)))
	protected.Handle("PUT /api/v1/products/{id}", requirePerm("products.update")(http.HandlerFunc(productH.Update)))
	protected.Handle("DELETE /api/v1/products/{id}", requirePerm("products.delete")(http.HandlerFunc(productH.Delete)))

	// Categories → categories.*
	protected.Handle("GET /api/v1/categories", requirePerm("categories.read")(http.HandlerFunc(productH.ListCategories)))
	protected.Handle("POST /api/v1/categories", requirePerm("categories.create")(http.HandlerFunc(productH.CreateCategory)))

	// Transactions → transactions.*
	protected.Handle("GET /api/v1/transactions", requirePerm("transactions.read")(http.HandlerFunc(txH.List)))
	protected.Handle("GET /api/v1/transactions/{id}", requirePerm("transactions.read")(http.HandlerFunc(txH.GetByID)))
	protected.Handle("POST /api/v1/transactions/checkout", requirePerm("transactions.create")(http.HandlerFunc(txH.Checkout)))

	// Stock / Inventory → stock.*
	protected.Handle("POST /api/v1/branches/{id}/inventory/adjustment", requirePerm("stock.adjust")(http.HandlerFunc(stockH.Adjustment)))
	protected.Handle("POST /api/v1/inventory/transfer", requirePerm("stock.transfer")(http.HandlerFunc(stockH.Transfer)))
	protected.Handle("GET /api/v1/branches/{id}/inventory", requirePerm("stock.read")(http.HandlerFunc(stockH.ListInventory)))
	protected.Handle("GET /api/v1/branches/{id}/inventory/low-stock", requirePerm("stock.read")(http.HandlerFunc(stockH.LowStock)))

	// Reports → reports.*
	protected.Handle("GET /api/v1/branches/{id}/reports/sales", requirePerm("reports.sales")(http.HandlerFunc(reportH.Sales)))
	protected.Handle("GET /api/v1/branches/{id}/reports/sales.pdf", requirePerm("reports.sales")(http.HandlerFunc(reportH.SalesPDF)))
	protected.Handle("GET /api/v1/branches/{id}/reports/stock", requirePerm("reports.stock")(http.HandlerFunc(reportH.Stock)))
	protected.Handle("GET /api/v1/branches/{id}/reports/profit-loss", requirePerm("reports.profit-loss")(http.HandlerFunc(reportH.ProfitLoss)))

	// Export → reports.sales
	protected.Handle("GET /api/v1/branches/{id}/reports/sales/export", requirePerm("reports.sales")(http.HandlerFunc(exportH.SalesExport)))

	// Users management → users.*
	protected.Handle("GET /api/v1/users", requirePerm("users.read")(http.HandlerFunc(userH.List)))
	protected.Handle("GET /api/v1/users/{id}", requirePerm("users.read")(http.HandlerFunc(userH.GetByID)))
	protected.Handle("POST /api/v1/users", requirePerm("users.create")(http.HandlerFunc(userH.Create)))
	protected.Handle("PUT /api/v1/users/{id}", requirePerm("users.update")(http.HandlerFunc(userH.Update)))
	protected.Handle("DELETE /api/v1/users/{id}", requirePerm("users.delete")(http.HandlerFunc(userH.Delete)))

	// Dashboard → dashboard.*
	protected.Handle("GET /api/v1/dashboard/stats", requirePerm("dashboard.stats")(http.HandlerFunc(dashboardH.DashboardStats)))
	protected.Handle("GET /api/v1/dashboard/sales-chart", requirePerm("dashboard.sales-chart")(http.HandlerFunc(dashboardH.SalesChart)))

	// Roles management → roles.*
	protected.Handle("GET /api/v1/roles", requirePerm("roles.read")(http.HandlerFunc(roleH.List)))
	protected.Handle("GET /api/v1/roles/{id}", requirePerm("roles.read")(http.HandlerFunc(roleH.GetByID)))
	protected.Handle("POST /api/v1/roles", requirePerm("roles.create")(http.HandlerFunc(roleH.Create)))
	protected.Handle("PUT /api/v1/roles/{id}", requirePerm("roles.update")(http.HandlerFunc(roleH.Update)))
	protected.Handle("DELETE /api/v1/roles/{id}", requirePerm("roles.delete")(http.HandlerFunc(roleH.Delete)))
	protected.Handle("GET /api/v1/roles/permissions/list", requirePerm("roles.read")(http.HandlerFunc(roleH.PermissionsList)))
	protected.Handle("GET /api/v1/roles/{id}/permissions", requirePerm("roles.read")(http.HandlerFunc(roleH.GetPermissions)))
	protected.Handle("PUT /api/v1/roles/{id}/permissions", requirePerm("roles.update")(http.HandlerFunc(roleH.SetPermissions)))

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
