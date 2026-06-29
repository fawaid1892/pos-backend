package database

import (
	"fmt"
	"log"
	"time"

	"pos-multi-branch/backend/internal/model"

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
