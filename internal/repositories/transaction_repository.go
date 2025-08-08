package repositories

import (
	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("Wallet").Preload("TransactionType").First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByReference(reference string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("Wallet").Preload("TransactionType").
		Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByWalletID(walletID uint, offset, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("TransactionType").
		Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) CalculateBalance(walletID uint) (decimal.Decimal, error) {
	var creditSum decimal.Decimal
	var debitSum decimal.Decimal

	// Calculate sum of credits (CREDIT transactions)
	err := r.db.Table("transactions t").
		Joins("JOIN transaction_types tt ON t.transaction_type_id = tt.id").
		Where("t.wallet_id = ? AND t.status = ? AND tt.name = ?",
			walletID, models.TransactionStatusCompleted, models.TransactionTypeCredit).
		Select("COALESCE(SUM(t.amount), 0)").
		Scan(&creditSum).Error

	if err != nil {
		return decimal.Zero, err
	}

	// Calculate sum of debits (DEBIT transactions)
	err = r.db.Table("transactions t").
		Joins("JOIN transaction_types tt ON t.transaction_type_id = tt.id").
		Where("t.wallet_id = ? AND t.status = ? AND tt.name = ?",
			walletID, models.TransactionStatusCompleted, models.TransactionTypeDebit).
		Select("COALESCE(SUM(t.amount), 0)").
		Scan(&debitSum).Error
	if err != nil {
		return decimal.Zero, err
	}

	return creditSum.Sub(debitSum), nil
}

func (r *transactionRepository) List(offset, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("Wallet").Preload("TransactionType").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	return transactions, err
}
