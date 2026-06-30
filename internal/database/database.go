package database

import (
	"fmt"
	"log"
	"time"

	"pos-multi-branch/backend/internal/idgen"
	"pos-multi-branch/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(databaseURL string) error {
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  databaseURL,
		PreferSimpleProtocol: true, // Disable prepared statements for PgBouncer
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open gorm connection: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected via GORM")
	return nil
}

func Migrate() error {
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Branch{},
		&model.Category{},
		&model.Product{},
		&model.Transaction{},
		&model.TransactionItem{},
		&model.BranchProduct{},
		&model.StockMutation{},
		&model.RefreshToken{},
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.Promotion{},
		&model.PromotionBranch{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	log.Println("AutoMigrate completed")

	// Seed default data if no users exist
	if err := Seed(); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}

	return nil
}

func Seed() error {
	log.Println("Seed: overwriting existing data...")

	// Delete existing data in reverse dependency order
	DB.Exec("DELETE FROM transaction_items")
	DB.Exec("DELETE FROM transactions")
	DB.Exec("DELETE FROM stock_mutations")
	DB.Exec("DELETE FROM branch_products")
	DB.Exec("DELETE FROM products")
	DB.Exec("DELETE FROM categories")
	DB.Exec("DELETE FROM refresh_tokens")
	DB.Exec("DELETE FROM role_permissions")
	DB.Exec("DELETE FROM users")
	DB.Exec("DELETE FROM promotions")
	DB.Exec("DELETE FROM promotion_branches")
	DB.Exec("DELETE FROM roles")
	DB.Exec("DELETE FROM permissions")
	DB.Exec("DELETE FROM branches")

	// Hash the default password
	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		return fmt.Errorf("bcrypt hash: %w", err)
	}
	password := string(hashed)

	// Create branches (manual ID)
	branchPusat := model.Branch{
		ID:           idgen.Generate(),
		Name:         "Cabang Pusat",
		Code:         "PST",
		Address:      "Jl. Merdeka No.1, Jakarta",
		Phone:        "021-1234567",
		Province:     "DKI Jakarta",
		City:         "Jakarta Pusat",
		ProvinceCode: "31",
		CityCode:     "3171",
		IsActive:     true,
	}
	branchCibubur := model.Branch{
		ID:           idgen.Generate(),
		Name:         "Cabang Cibubur",
		Code:         "CBG-01",
		Address:      "Jl. Cibubur Raya No.5, Bekasi",
		Phone:        "021-7654321",
		Province:     "Jawa Barat",
		City:         "Bekasi",
		ProvinceCode: "32",
		CityCode:     "3275",
		IsActive:     true,
	}
	if err := DB.Create(&branchPusat).Error; err != nil {
		return fmt.Errorf("seed branch pusat: %w", err)
	}
	if err := DB.Create(&branchCibubur).Error; err != nil {
		return fmt.Errorf("seed branch cibubur: %w", err)
	}
	log.Println("Seed: branches created")

	// Create users (using branch IDs from created branches)
	admin := model.User{
		ID:       idgen.Generate(),
		Username: "admin",
		Password: password,
		FullName: "Admin Utama",
		Role:     "admin_cabang",
		BranchID: &branchPusat.ID,
		IsActive: true,
	}
	kasir1 := model.User{
		ID:       idgen.Generate(),
		Username: "kasir1",
		Password: password,
		FullName: "Kasir Cabang 1",
		Role:     "kasir",
		BranchID: &branchPusat.ID,
		IsActive: true,
	}
	owner := model.User{
		ID:       idgen.Generate(),
		Username: "owner",
		Password: password,
		FullName: "Pemilik Toko",
		Role:     "owner",
		IsActive: true,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	if err := DB.Create(&kasir1).Error; err != nil {
		return fmt.Errorf("seed kasir1: %w", err)
	}
	if err := DB.Create(&owner).Error; err != nil {
		return fmt.Errorf("seed owner: %w", err)
	}
	log.Println("Seed: users created")

	// Create categories
	categories := []model.Category{
		{ID: idgen.Generate(), Name: "Makanan"},
		{ID: idgen.Generate(), Name: "Minuman"},
		{ID: idgen.Generate(), Name: "Snack"},
		{ID: idgen.Generate(), Name: "Alat Tulis"},
	}
	for _, cat := range categories {
		if err := DB.Create(&cat).Error; err != nil {
			return fmt.Errorf("seed category %s: %w", cat.Name, err)
		}
	}
	log.Println("Seed: categories created")

	// Look up category IDs for products
	var makanan, minuman, snack, alatTulis model.Category
	DB.Where("name = ?", "Makanan").First(&makanan)
	DB.Where("name = ?", "Minuman").First(&minuman)
	DB.Where("name = ?", "Snack").First(&snack)
	DB.Where("name = ?", "Alat Tulis").First(&alatTulis)

	// Create products
	barcode1 := "8991001001001"
	barcode2 := "8991001001002"
	barcode3 := "8991001001003"
	barcode4 := "8991001001004"
	products := []model.Product{
		{ID: idgen.Generate(), CategoryID: makanan.ID, Name: "Nasi Goreng", Barcode: &barcode1, Unit: "PCS", Price: 15000, Stock: 50},
		{ID: idgen.Generate(), CategoryID: minuman.ID, Name: "Air Mineral 600ml", Barcode: &barcode2, Unit: "PCS", Price: 5000, Stock: 100},
		{ID: idgen.Generate(), CategoryID: snack.ID, Name: "Keripik Singkong", Barcode: &barcode3, Unit: "PCS", Price: 8000, Stock: 75},
		{ID: idgen.Generate(), CategoryID: alatTulis.ID, Name: "Pulpen Standard", Barcode: &barcode4, Unit: "PCS", Price: 3000, Stock: 200},
	}
	for _, prod := range products {
		if err := DB.Create(&prod).Error; err != nil {
			return fmt.Errorf("seed product %s: %w", prod.Name, err)
		}
	}
	log.Println("Seed: products created")

	// ─── Seed Permissions ───
	permissions := []model.Permission{
		{ID: idgen.Generate(), Name: "products.read", Label: "Melihat Produk", Group: "products"},
		{ID: idgen.Generate(), Name: "products.create", Label: "Menambah Produk", Group: "products"},
		{ID: idgen.Generate(), Name: "products.update", Label: "Mengubah Produk", Group: "products"},
		{ID: idgen.Generate(), Name: "products.delete", Label: "Menghapus Produk", Group: "products"},
		{ID: idgen.Generate(), Name: "categories.read", Label: "Melihat Kategori", Group: "categories"},
		{ID: idgen.Generate(), Name: "categories.create", Label: "Menambah Kategori", Group: "categories"},
		{ID: idgen.Generate(), Name: "branches.read", Label: "Melihat Cabang", Group: "branches"},
		{ID: idgen.Generate(), Name: "branches.create", Label: "Menambah Cabang", Group: "branches"},
		{ID: idgen.Generate(), Name: "branches.update", Label: "Mengubah Cabang", Group: "branches"},
		{ID: idgen.Generate(), Name: "branches.delete", Label: "Menghapus Cabang", Group: "branches"},
		{ID: idgen.Generate(), Name: "users.read", Label: "Melihat Pengguna", Group: "users"},
		{ID: idgen.Generate(), Name: "users.create", Label: "Menambah Pengguna", Group: "users"},
		{ID: idgen.Generate(), Name: "users.update", Label: "Mengubah Pengguna", Group: "users"},
		{ID: idgen.Generate(), Name: "users.delete", Label: "Menghapus Pengguna", Group: "users"},
		{ID: idgen.Generate(), Name: "transactions.read", Label: "Melihat Transaksi", Group: "transactions"},
		{ID: idgen.Generate(), Name: "transactions.create", Label: "Membuat Transaksi", Group: "transactions"},
		{ID: idgen.Generate(), Name: "stock.read", Label: "Melihat Stok", Group: "stock"},
		{ID: idgen.Generate(), Name: "stock.adjust", Label: "Menyesuaikan Stok", Group: "stock"},
		{ID: idgen.Generate(), Name: "stock.transfer", Label: "Transfer Stok", Group: "stock"},
		{ID: idgen.Generate(), Name: "reports.sales", Label: "Laporan Penjualan", Group: "reports"},
		{ID: idgen.Generate(), Name: "reports.stock", Label: "Laporan Stok", Group: "reports"},
		{ID: idgen.Generate(), Name: "reports.profit-loss", Label: "Laporan Laba Rugi", Group: "reports"},
		{ID: idgen.Generate(), Name: "dashboard.stats", Label: "Statistik Dashboard", Group: "dashboard"},
		{ID: idgen.Generate(), Name: "dashboard.sales-chart", Label: "Grafik Penjualan", Group: "dashboard"},
		{ID: idgen.Generate(), Name: "settings.read", Label: "Melihat Pengaturan", Group: "settings"},
		{ID: idgen.Generate(), Name: "settings.update", Label: "Mengubah Pengaturan", Group: "settings"},
		{ID: idgen.Generate(), Name: "roles.read", Label: "Melihat Role", Group: "roles"},
		{ID: idgen.Generate(), Name: "roles.create", Label: "Menambah Role", Group: "roles"},
		{ID: idgen.Generate(), Name: "roles.update", Label: "Mengubah Role", Group: "roles"},
		{ID: idgen.Generate(), Name: "roles.delete", Label: "Menghapus Role", Group: "roles"},
		// Promotions
		{ID: idgen.Generate(), Name: "promotions.read", Label: "Melihat Promosi", Group: "promotions"},
		{ID: idgen.Generate(), Name: "promotions.create", Label: "Menambah Promosi", Group: "promotions"},
		{ID: idgen.Generate(), Name: "promotions.update", Label: "Mengubah Promosi", Group: "promotions"},
		{ID: idgen.Generate(), Name: "promotions.delete", Label: "Menghapus Promosi", Group: "promotions"},
	}

	permMap := make(map[string]int64)
	for _, p := range permissions {
		var existing model.Permission
		if err := DB.Where("name = ?", p.Name).First(&existing).Error; err != nil {
			DB.Create(&p)
			permMap[p.Name] = p.ID
		} else {
			permMap[p.Name] = existing.ID
		}
	}
	log.Println("Seed: permissions seeded")

	// ─── Seed Roles ───
	kasirRole := model.Role{ID: idgen.Generate(), Name: "kasir", Description: "Kasir — melayani transaksi", IsSystem: true}
	adminRole := model.Role{ID: idgen.Generate(), Name: "admin_cabang", Description: "Admin cabang — mengelola operasional cabang", IsSystem: true}
	ownerRole := model.Role{ID: idgen.Generate(), Name: "owner", Description: "Pemilik — akses penuh", IsSystem: true}

	var existingKasir, existingAdmin, existingOwner model.Role
	DB.Where("name = ?", "kasir").First(&existingKasir)
	DB.Where("name = ?", "admin_cabang").First(&existingAdmin)
	DB.Where("name = ?", "owner").First(&existingOwner)

	kasirRoleID := existingKasir.ID
	adminRoleID := existingAdmin.ID
	ownerRoleID := existingOwner.ID

	if kasirRoleID == 0 {
		DB.Create(&kasirRole)
		kasirRoleID = kasirRole.ID
	}
	if adminRoleID == 0 {
		DB.Create(&adminRole)
		adminRoleID = adminRole.ID
	}
	if ownerRoleID == 0 {
		DB.Create(&ownerRole)
		ownerRoleID = ownerRole.ID
	}
	log.Println("Seed: roles seeded")

	// ─── Assign Permissions ───
	// kasir: products.read, categories.read, transactions.read, transactions.create, stock.read, stock.adjust
	kasirPerms := []string{"products.read", "categories.read", "transactions.read", "transactions.create", "stock.read", "stock.adjust"}

	// Clear old permissions
	DB.Where("role_id IN (?, ?, ?)", kasirRoleID, adminRoleID, ownerRoleID).Delete(&model.RolePermission{})

	// Helper to assign perms
	assignPerms := func(roleID int64, permNames []string) {
		for _, name := range permNames {
			if pid, ok := permMap[name]; ok {
				DB.Create(&model.RolePermission{RoleID: roleID, PermissionID: pid})
			}
		}
	}

	assignPerms(kasirRoleID, kasirPerms)

	// admin_cabang: all except roles.*
	var adminPerms []string
	for name := range permMap {
		if name != "roles.read" && name != "roles.create" && name != "roles.update" && name != "roles.delete" {
			adminPerms = append(adminPerms, name)
		}
	}
	assignPerms(adminRoleID, adminPerms)

	// owner: all
	var allPerms []string
	for name := range permMap {
		allPerms = append(allPerms, name)
	}
	assignPerms(ownerRoleID, allPerms)

	log.Println("Seed: role permissions assigned")

	// ─── Assign roles to seed users ───
	DB.Model(&model.User{}).Where("username = ?", "admin").UpdateColumn("role_id", adminRoleID)
	DB.Model(&model.User{}).Where("username = ?", "kasir1").UpdateColumn("role_id", kasirRoleID)
	DB.Model(&model.User{}).Where("username = ?", "owner").UpdateColumn("role_id", ownerRoleID)
	log.Println("Seed: role IDs assigned to seed users")

	log.Println("Seed: default data seeded successfully")
	return nil
}

func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
