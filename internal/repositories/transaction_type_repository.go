package repositories

import (
	"github.com/limistah/wallet-service/internal/models"
	"gorm.io/gorm"
)

type transactionTypeRepository struct {
	db *gorm.DB
}

// NewTransactionTypeRepository creates a new transaction type repository
func NewTransactionTypeRepository(db *gorm.DB) TransactionTypeRepository {
	return &transactionTypeRepository{db: db}
}

func (r *transactionTypeRepository) GetByName(name string) (*models.TransactionType, error) {
	var transactionType models.TransactionType
	err := r.db.Where("name = ?", name).First(&transactionType).Error
	if err != nil {
		return nil, err
	}
	return &transactionType, nil
}

func (r *transactionTypeRepository) List() ([]models.TransactionType, error) {
	var transactionTypes []models.TransactionType
	err := r.db.Find(&transactionTypes).Error
	return transactionTypes, err
}

func (r *transactionTypeRepository) Create(transactionType *models.TransactionType) error {
	return r.db.Create(transactionType).Error
}
