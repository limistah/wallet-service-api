package database

import (
	"fmt"
	"log"
	"os"

	"github.com/limistah/wallet-service/internal/config"
	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize connects to database and runs migrations
func Initialize() (*gorm.DB, error) {
	cfg := config.LoadConfig()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}

	var db *gorm.DB
	var err error

	gormLogger := logger.Default
	if cfg.App.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	switch cfg.Database.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL database: %v", err)
		}
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if cfg.Database.Driver == "mysql" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %v", err)
		}

		sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %v", err)
		}
	}

	log.Printf("Successfully connected to %s database", cfg.Database.Driver)

	err = db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.ReconciliationReport{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	err = bootstrapSystemAccount(db)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap system account: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

// InitWithConfig initializes database with provided config (useful for testing)
func InitWithConfig(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormLogger := logger.Default
	if cfg.App.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	switch cfg.Database.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(":memory:"), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %v", cfg.Database.Driver, err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.ReconciliationReport{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

// bootstrapSystemAccount creates the system account and wallet for double-entry bookkeeping
func bootstrapSystemAccount(db *gorm.DB) error {
	// Check if system account already exists
	var existingUser models.User
	if err := db.Where("email = ? AND is_system = ?", models.SystemAccountEmail, true).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			systemUser := models.CreateSystemUser()

			if err := systemUser.HashPassword(systemUser.Password); err != nil {
				return fmt.Errorf("failed to hash system account password: %v", err)
			}

			if err := db.Create(systemUser).Error; err != nil {
				return fmt.Errorf("failed to create system user: %v", err)
			}

			systemWallet := &models.Wallet{
				UserID:   systemUser.ID,
				Balance:  decimal.NewFromInt(1000000000), // 1 billion as initial system balance
				Currency: "USD",
				Status:   models.WalletStatusActive,
			}

			if err := db.Create(systemWallet).Error; err != nil {
				return fmt.Errorf("failed to create system wallet: %v", err)
			}

			log.Printf("System account and wallet created successfully with ID: %d", systemUser.ID)
		} else {
			return fmt.Errorf("failed to check for existing system account: %v", err)
		}
	} else {
		log.Printf("System account already exists with ID: %d", existingUser.ID)
	}

	return nil
}
