package usecases

import (
	"fmt"
	"time"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/shopspring/decimal"
)

// SystemReconciliationReport represents system-wide reconciliation results
type SystemReconciliationReport struct {
	TotalWallets       int             `json:"total_wallets"`
	WalletsWithIssues  int             `json:"wallets_with_issues"`
	SystemBalance      decimal.Decimal `json:"system_balance"`
	GlobalBalanceCheck decimal.Decimal `json:"global_balance_check"`
	Issues             []string        `json:"issues"`
	CreatedAt          time.Time       `json:"created_at"`
}

// SystemAccountValidation represents system account validation results
type SystemAccountValidation struct {
	SystemWalletID  uint            `json:"system_wallet_id"`
	StoredBalance   decimal.Decimal `json:"stored_balance"`
	ExpectedBalance decimal.Decimal `json:"expected_balance"`
	InitialBalance  decimal.Decimal `json:"initial_balance"`
	Difference      decimal.Decimal `json:"difference"`
	IsValid         bool            `json:"is_valid"`
	Issues          []string        `json:"issues"`
}

type reconciliationUseCase struct {
	repos *repositories.Repositories
}

// NewReconciliationUseCase creates a new reconciliation use case
func NewReconciliationUseCase(repos *repositories.Repositories) ReconciliationUseCase {
	return &reconciliationUseCase{repos: repos}
}

func (uc *reconciliationUseCase) PerformReconciliation() ([]models.ReconciliationReport, error) {
	// Get all wallets for reconciliation
	wallets, err := uc.repos.Wallet.GetAllForReconciliation()
	if err != nil {
		return nil, err
	}

	var reports []models.ReconciliationReport

	for _, wallet := range wallets {
		report, err := uc.performWalletReconciliation(wallet.ID)
		if err != nil {
			// Log error but continue with other wallets
			continue
		}
		reports = append(reports, *report)
	}

	return reports, nil
}

func (uc *reconciliationUseCase) PerformWalletReconciliation(walletID uint) (*models.ReconciliationReport, error) {
	return uc.performWalletReconciliation(walletID)
}

func (uc *reconciliationUseCase) performWalletReconciliation(walletID uint) (*models.ReconciliationReport, error) {
	// Get wallet
	wallet, err := uc.repos.Wallet.GetByID(walletID)
	if err != nil {
		return nil, err
	}

	// Calculate balance from transactions
	calculatedBalance, err := uc.repos.Transaction.CalculateBalance(walletID)
	if err != nil {
		return nil, err
	}

	// Compare balances
	storedBalance := wallet.Balance
	difference := storedBalance.Sub(calculatedBalance)

	// Determine status
	status := models.ReconciliationStatusMatch
	notes := "Balance matches"

	if !difference.IsZero() {
		status = models.ReconciliationStatusMismatch
		notes = fmt.Sprintf("Balance mismatch detected. Difference: %s", difference.String())
	}

	// Create reconciliation report
	report := &models.ReconciliationReport{
		WalletID:          walletID,
		StoredBalance:     storedBalance,
		CalculatedBalance: calculatedBalance,
		Difference:        difference,
		Status:            status,
		Notes:             notes,
	}

	// Save the report
	err = uc.repos.Reconciliation.Create(report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (uc *reconciliationUseCase) GetReconciliationReports(page, pageSize int) ([]models.ReconciliationReport, error) {
	offset := (page - 1) * pageSize
	return uc.repos.Reconciliation.List(offset, pageSize)
}

func (uc *reconciliationUseCase) GetMismatchReports(page, pageSize int) ([]models.ReconciliationReport, error) {
	offset := (page - 1) * pageSize
	return uc.repos.Reconciliation.GetMismatches(offset, pageSize)
}
