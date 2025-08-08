package repositories

import (
	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) GetByID(id uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Preload("User").First(&wallet, id).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Update(wallet *models.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *walletRepository) UpdateBalance(walletID uint, newBalance decimal.Decimal, version uint) error {
	// Optimistic locking: update only if version matches
	result := r.db.Model(&models.Wallet{}).
		Where("id = ? AND version = ?", walletID, version).
		Updates(map[string]interface{}{
			"balance": newBalance,
			"version": version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Version mismatch or record not found
	}

	return nil
}

func (r *walletRepository) List(offset, limit int) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Preload("User").Offset(offset).Limit(limit).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) GetAllForReconciliation() ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Preload("User").Find(&wallets).Error
	return wallets, err
}
