package database

import (
	"fmt"
	"log"
	"time"

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
	DB.Exec("DELETE FROM users")
	DB.Exec("DELETE FROM branches")

	// Hash the default password
	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		return fmt.Errorf("bcrypt hash: %w", err)
	}
	password := string(hashed)

	// Create branches
	branchPusat := model.Branch{
		Name:    "Cabang Pusat",
		Address: "Jl. Merdeka No.1, Jakarta",
		Phone:   "021-1234567",
		IsActive: true,
	}
	branchCibubur := model.Branch{
		Name:    "Cabang Cibubur",
		Address: "Jl. Cibubur Raya No.5, Bekasi",
		Phone:   "021-7654321",
		IsActive: true,
	}
	if err := DB.Create(&branchPusat).Error; err != nil {
		return fmt.Errorf("seed branch pusat: %w", err)
	}
	if err := DB.Create(&branchCibubur).Error; err != nil {
		return fmt.Errorf("seed branch cibubur: %w", err)
	}
	log.Println("Seed: branches created")

	// Create users (using branch IDs)
	admin := model.User{
		Username: "admin",
		Password: password,
		FullName: "Admin Utama",
		Role:     "admin_cabang",
		BranchID: &branchPusat.ID,
		IsActive: true,
	}
	kasir1 := model.User{
		Username: "kasir1",
		Password: password,
		FullName: "Kasir Cabang 1",
		Role:     "kasir",
		BranchID: &branchPusat.ID,
		IsActive: true,
	}
	owner := model.User{
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
		{Name: "Makanan"},
		{Name: "Minuman"},
		{Name: "Snack"},
		{Name: "Alat Tulis"},
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
	products := []model.Product{
		{CategoryID: makanan.ID, Name: "Nasi Goreng", Barcode: "8991001001001", Price: 15000, Stock: 50},
		{CategoryID: minuman.ID, Name: "Air Mineral 600ml", Barcode: "8991001001002", Price: 5000, Stock: 100},
		{CategoryID: snack.ID, Name: "Keripik Singkong", Barcode: "8991001001003", Price: 8000, Stock: 75},
		{CategoryID: alatTulis.ID, Name: "Pulpen Standard", Barcode: "8991001001004", Price: 3000, Stock: 200},
	}
	for _, prod := range products {
		if err := DB.Create(&prod).Error; err != nil {
			return fmt.Errorf("seed product %s: %w", prod.Name, err)
		}
	}
	log.Println("Seed: products created")

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
