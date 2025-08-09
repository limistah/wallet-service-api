package repositories

import (
	"time"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]models.User, error)
}

// WalletRepository defines the interface for wallet data operations
type WalletRepository interface {
	Create(wallet *models.Wallet) error
	GetByID(id uint) (*models.Wallet, error)
	GetByUserID(userID uint) (*models.Wallet, error)
	Update(wallet *models.Wallet) error
	UpdateBalance(walletID uint, newBalance decimal.Decimal, version uint) error
	List(offset, limit int) ([]models.Wallet, error)
	GetAllForReconciliation() ([]models.Wallet, error)
}

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id uint) (*models.Transaction, error)
	GetByReference(reference string) (*models.Transaction, error)
	GetByWalletID(walletID uint, offset, limit int) ([]models.Transaction, error)
	GetByWalletIDWithCursor(walletID uint, cursor *time.Time, cursorID *uint, limit int) ([]models.Transaction, error)
	Update(transaction *models.Transaction) error
	CalculateBalance(walletID uint) (decimal.Decimal, error)
	List(offset, limit int) ([]models.Transaction, error)
}

// TransactionTypeRepository defines the interface for transaction type operations
type TransactionTypeRepository interface {
	GetByName(name string) (*models.TransactionType, error)
	List() ([]models.TransactionType, error)
	Create(transactionType *models.TransactionType) error
}

// ReconciliationRepository defines the interface for reconciliation operations
type ReconciliationRepository interface {
	Create(report *models.ReconciliationReport) error
	GetByWalletID(walletID uint) ([]models.ReconciliationReport, error)
	List(offset, limit int) ([]models.ReconciliationReport, error)
	GetMismatches(offset, limit int) ([]models.ReconciliationReport, error)
}

// Repositories holds all repository interfaces
type Repositories struct {
	User            UserRepository
	Wallet          WalletRepository
	Transaction     TransactionRepository
	TransactionType TransactionTypeRepository
	Reconciliation  ReconciliationRepository
	DB              *gorm.DB
}

// NewRepositories creates a new instance of all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:           NewUserRepository(db),
		Wallet:         NewWalletRepository(db),
		Transaction:    NewTransactionRepository(db),
		Reconciliation: NewReconciliationRepository(db),
		DB:             db,
	}
}
