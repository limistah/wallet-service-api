package repositories

import (
	"github.com/limistah/wallet-service/internal/models"
	"gorm.io/gorm"
)

type reconciliationRepository struct {
	db *gorm.DB
}

// NewReconciliationRepository creates a new reconciliation repository
func NewReconciliationRepository(db *gorm.DB) ReconciliationRepository {
	return &reconciliationRepository{db: db}
}

func (r *reconciliationRepository) Create(report *models.ReconciliationReport) error {
	return r.db.Create(report).Error
}

func (r *reconciliationRepository) GetByWalletID(walletID uint) ([]models.ReconciliationReport, error) {
	var reports []models.ReconciliationReport
	err := r.db.Preload("Wallet").
		Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *reconciliationRepository) List(offset, limit int) ([]models.ReconciliationReport, error) {
	var reports []models.ReconciliationReport
	err := r.db.Preload("Wallet").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reports).Error
	return reports, err
}

func (r *reconciliationRepository) GetMismatches(offset, limit int) ([]models.ReconciliationReport, error) {
	var reports []models.ReconciliationReport
	err := r.db.Preload("Wallet").
		Where("status = ?", models.ReconciliationStatusMismatch).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reports).Error
	return reports, err
}
