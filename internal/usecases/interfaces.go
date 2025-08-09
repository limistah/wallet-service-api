package usecases

import (
	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/shopspring/decimal"
)

// UserUseCase defines the interface for user business logic
type UserUseCase interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUser(id uint) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(id uint, user *models.User) (*models.User, error)
	DeleteUser(id uint) error
	ListUsers(page, pageSize int) ([]models.User, error)
}

// WalletUseCase defines the interface for wallet business logic
type WalletUseCase interface {
	CreateWallet(userID uint, currency string) (*models.Wallet, error)
	GetWallet(id uint) (*models.Wallet, error)
	GetWalletByUserID(userID uint) (*models.Wallet, error)
	FundWallet(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error)
	WithdrawFunds(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error)
	TransferFunds(fromWalletID, toWalletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error)
	GetWalletBalance(walletID uint) (decimal.Decimal, error)
	GetTransactionHistory(walletID uint, cursor *string, limit int) ([]models.Transaction, *string, error)
}

// ReconciliationUseCase defines the interface for reconciliation business logic
type ReconciliationUseCase interface {
	PerformReconciliation() ([]models.ReconciliationReport, error)
	PerformWalletReconciliation(walletID uint) (*models.ReconciliationReport, error)
	GetReconciliationReports(page, pageSize int) ([]models.ReconciliationReport, error)
	GetMismatchReports(page, pageSize int) ([]models.ReconciliationReport, error)
}

// UseCases holds all use case interfaces
type UseCases struct {
	User           UserUseCase
	Wallet         WalletUseCase
	Reconciliation ReconciliationUseCase
}

// NewUseCases creates a new instance of all use cases
func NewUseCases(repos *repositories.Repositories) *UseCases {
	reconciliationUC := NewReconciliationUseCase(repos)

	return &UseCases{
		User:           NewUserUseCase(repos),
		Wallet:         NewWalletUseCase(repos, reconciliationUC),
		Reconciliation: reconciliationUC,
	}
}
