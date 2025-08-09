package usecases

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type walletUseCase struct {
	repos            *repositories.Repositories
	reconciliationUC ReconciliationUseCase
}

// TransactionCursor represents a cursor for pagination
type TransactionCursor struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewWalletUseCase creates a new wallet use case
func NewWalletUseCase(repos *repositories.Repositories, reconciliationUC ReconciliationUseCase) WalletUseCase {
	return &walletUseCase{
		repos:            repos,
		reconciliationUC: reconciliationUC,
	}
}

// performPreTransactionReconciliation performs reconciliation check before withdrawal/transfer
// This ensures the wallet balance is accurate before any debiting operation
func (uc *walletUseCase) performPreTransactionReconciliation(walletID uint) error {
	report, err := uc.reconciliationUC.PerformWalletReconciliation(walletID)
	if err != nil {
		return fmt.Errorf("reconciliation check failed: %w", err)
	}

	if report.Status == models.ReconciliationStatusMismatch {
		return fmt.Errorf("wallet balance mismatch detected: stored=%s, calculated=%s, difference=%s. Transaction cannot proceed until reconciliation is resolved",
			report.StoredBalance.String(), report.CalculatedBalance.String(), report.Difference.String())
	}

	return nil
}

// performPostTransactionReconciliation performs reconciliation after transaction for audit
// This is optional and won't block transactions, but provides audit trail
func (uc *walletUseCase) performPostTransactionReconciliation(walletID uint) {
	// This is for audit purposes only
	err := uc.performPreTransactionReconciliation(walletID)
	if err != nil {
		fmt.Printf("Post-transaction reconciliation warning for wallet %d: %v\n", walletID, err)
	}
}

// getSystemWallet retrieves the system wallet for double-entry bookkeeping
func (uc *walletUseCase) getSystemWallet() (*models.Wallet, error) {
	systemUser, err := uc.repos.User.GetByEmail(models.SystemAccountEmail)
	if err != nil {
		return nil, fmt.Errorf("system user not found: %w", err)
	}

	systemWallet, err := uc.repos.Wallet.GetByUserID(systemUser.ID)
	if err != nil {
		return nil, fmt.Errorf("system wallet not found: %w", err)
	}

	return systemWallet, nil
}

func (uc *walletUseCase) CreateWallet(userID uint, currency string) (*models.Wallet, error) {
	_, err := uc.repos.User.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	existingWallet, err := uc.repos.Wallet.GetByUserID(userID)
	if err == nil && existingWallet != nil {
		return nil, errors.New("user already has a wallet")
	}

	wallet := &models.Wallet{
		UserID:   userID,
		Balance:  decimal.Zero,
		Currency: currency,
		Status:   models.WalletStatusActive,
	}

	err = uc.repos.Wallet.Create(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (uc *walletUseCase) GetWallet(id uint) (*models.Wallet, error) {
	return uc.repos.Wallet.GetByID(id)
}

func (uc *walletUseCase) GetWalletByUserID(userID uint) (*models.Wallet, error) {
	return uc.repos.Wallet.GetByUserID(userID)
}

func (uc *walletUseCase) FundWallet(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("amount must be greater than zero")
	}

	if err := uc.performPreTransactionReconciliation(walletID); err != nil {
		return nil, nil, fmt.Errorf("pre-transaction reconciliation failed: %w", err)
	}

	_, err := uc.repos.Transaction.GetByReference(reference)
	if err == nil {
		return nil, nil, errors.New("duplicate reference")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, nil, fmt.Errorf("error checking reference: %w", err)
	}

	userWallet, err := uc.repos.Wallet.GetByID(walletID)
	if err != nil {
		return nil, nil, errors.New("wallet not found")
	}

	if !userWallet.IsActive() {
		return nil, nil, errors.New("wallet is not active")
	}

	systemWallet, err := uc.getSystemWallet()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get system wallet: %w", err)
	}

	if !systemWallet.CanDebit(amount) {
		return nil, nil, errors.New("insufficient system funds")
	}

	var systemTransaction, userTransaction *models.Transaction

	err = uc.repos.DB.Transaction(func(tx *gorm.DB) error {
		systemBalanceBefore := systemWallet.Balance
		systemBalanceAfter := systemBalanceBefore.Sub(amount)

		systemTransaction = &models.Transaction{
			Reference:          reference + "_system_debit",
			WalletID:           systemWallet.ID,
			TransactionType:    models.TransactionTypeDebit,
			Amount:             amount,
			Metadata:           `{"source": "funding"}`,
			BalanceBefore:      systemBalanceBefore,
			BalanceAfter:       systemBalanceAfter,
			TransactionPurpose: "WALLET_TOP_UP",
			Description:        fmt.Sprintf("System debit for funding: %s", description),
			Status:             models.TransactionStatusCompleted,
		}

		if err := tx.Create(systemTransaction).Error; err != nil {
			return fmt.Errorf("failed to create system transaction: %w", err)
		}

		result := tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", systemWallet.ID, systemWallet.Version).
			Updates(map[string]interface{}{
				"balance": systemBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update system wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("system wallet version mismatch - concurrent modification detected")
		}

		userBalanceBefore := userWallet.Balance
		userBalanceAfter := userBalanceBefore.Add(amount)

		userTransaction = &models.Transaction{
			Reference:            reference,
			WalletID:             walletID,
			TransactionType:      models.TransactionTypeCredit,
			Amount:               amount,
			Metadata:             `{"source": "funding"}`,
			BalanceBefore:        userBalanceBefore,
			BalanceAfter:         userBalanceAfter,
			TransactionPurpose:   "WALLET_TOP_UP",
			Description:          description,
			Status:               models.TransactionStatusCompleted,
			RelatedTransactionID: &systemTransaction.ID,
		}

		if err := tx.Create(userTransaction).Error; err != nil {
			return fmt.Errorf("failed to create user transaction: %w", err)
		}

		result = tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", walletID, userWallet.Version).
			Updates(map[string]interface{}{
				"balance": userBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update user wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("user wallet version mismatch - concurrent modification detected")
		}

		return tx.Model(systemTransaction).Update("related_transaction_id", userTransaction.ID).Error
	})

	if err != nil {
		return nil, nil, err
	}

	go uc.performPostTransactionReconciliation(walletID)

	userTx, err := uc.repos.Transaction.GetByID(userTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load user transaction: %w", err)
	}

	systemTx, err := uc.repos.Transaction.GetByID(systemTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system transaction: %w", err)
	}

	return userTx, systemTx, nil
}

func (uc *walletUseCase) WithdrawFunds(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("amount must be greater than zero")
	}

	if err := uc.performPreTransactionReconciliation(walletID); err != nil {
		return nil, nil, fmt.Errorf("pre-transaction reconciliation failed: %w", err)
	}

	_, err := uc.repos.Transaction.GetByReference(reference)
	if err == nil {
		return nil, nil, errors.New("duplicate reference")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, nil, fmt.Errorf("error checking reference: %w", err)
	}

	userWallet, err := uc.repos.Wallet.GetByID(walletID)
	if err != nil {
		return nil, nil, errors.New("wallet not found")
	}

	if !userWallet.IsActive() {
		return nil, nil, errors.New("wallet is not active")
	}

	if !userWallet.CanDebit(amount) {
		return nil, nil, fmt.Errorf("insufficient funds: available=%.2f, requested=%.2f",
			userWallet.Balance.InexactFloat64(), amount.InexactFloat64())
	}

	systemWallet, err := uc.getSystemWallet()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get system wallet: %w", err)
	}

	var userTransaction, systemTransaction *models.Transaction

	err = uc.repos.DB.Transaction(func(tx *gorm.DB) error {
		userBalanceBefore := userWallet.Balance
		userBalanceAfter := userBalanceBefore.Sub(amount)

		if userBalanceAfter.LessThan(decimal.Zero) {
			return errors.New("insufficient funds for withdrawal")
		}

		userTransaction = &models.Transaction{
			Reference:          reference,
			WalletID:           walletID,
			TransactionType:    models.TransactionTypeDebit,
			Amount:             amount,
			Metadata:           `{"source": "withdrawal"}`,
			BalanceBefore:      userBalanceBefore,
			BalanceAfter:       userBalanceAfter,
			TransactionPurpose: "WITHDRAWAL",
			Description:        description,
			Status:             models.TransactionStatusCompleted,
		}

		if err := tx.Create(userTransaction).Error; err != nil {
			return fmt.Errorf("failed to create user transaction: %w", err)
		}

		result := tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", walletID, userWallet.Version).
			Updates(map[string]interface{}{
				"balance": userBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update user wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("user wallet version mismatch - concurrent modification detected")
		}

		systemBalanceBefore := systemWallet.Balance
		systemBalanceAfter := systemBalanceBefore.Add(amount)

		systemTransaction = &models.Transaction{
			Reference:            reference + "_system_credit",
			WalletID:             systemWallet.ID,
			TransactionType:      models.TransactionTypeCredit,
			Amount:               amount,
			Metadata:             `{"source": "withdrawal"}`,
			BalanceBefore:        systemBalanceBefore,
			BalanceAfter:         systemBalanceAfter,
			TransactionPurpose:   "WITHDRAWAL",
			Description:          fmt.Sprintf("System credit for withdrawal: %s", description),
			Status:               models.TransactionStatusCompleted,
			RelatedTransactionID: &userTransaction.ID,
		}

		if err := tx.Create(systemTransaction).Error; err != nil {
			return fmt.Errorf("failed to create system transaction: %w", err)
		}

		result = tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", systemWallet.ID, systemWallet.Version).
			Updates(map[string]interface{}{
				"balance": systemBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update system wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("system wallet version mismatch - concurrent modification detected")
		}

		return tx.Model(userTransaction).Update("related_transaction_id", systemTransaction.ID).Error
	})

	if err != nil {
		return nil, nil, err
	}

	go uc.performPostTransactionReconciliation(walletID)

	userTx, err := uc.repos.Transaction.GetByID(userTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load user transaction: %w", err)
	}

	systemTx, err := uc.repos.Transaction.GetByID(systemTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system transaction: %w", err)
	}

	return userTx, systemTx, nil
}

func (uc *walletUseCase) TransferFunds(fromWalletID, toWalletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	// Validate different wallets
	if fromWalletID == toWalletID {
		return nil, nil, errors.New("cannot transfer to the same wallet")
	}
	// Get both wallets
	fromWallet, err := uc.repos.Wallet.GetByID(fromWalletID)
	if err != nil {
		return nil, nil, errors.New("source wallet not found")
	}

	toWallet, err := uc.repos.Wallet.GetByID(toWalletID)
	if err != nil {
		return nil, nil, errors.New("destination wallet not found")
	}

	if !fromWallet.CanDebit(amount) {
		return nil, nil, fmt.Errorf("insufficient funds in source wallet: available=%.2f, requested=%.2f",
			fromWallet.Balance.InexactFloat64(), amount.InexactFloat64())
	}

	if !toWallet.IsActive() {
		return nil, nil, errors.New("destination wallet is not active")
	}

	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("amount must be greater than zero")
	}

	if err := uc.performPreTransactionReconciliation(fromWalletID); err != nil {
		return nil, nil, fmt.Errorf("source wallet reconciliation failed: %w", err)
	}
	if err := uc.performPreTransactionReconciliation(toWalletID); err != nil {
		return nil, nil, fmt.Errorf("destination wallet reconciliation failed: %w", err)
	}

	// Check for duplicate reference
	_, err = uc.repos.Transaction.GetByReference(reference)
	if err == nil {
		return nil, nil, errors.New("duplicate reference")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, nil, fmt.Errorf("error checking reference: %w", err)
	}

	// Prevent transfers to system accounts (unless explicitly allowed)
	systemWallet, _ := uc.getSystemWallet()
	if systemWallet != nil && toWalletID == systemWallet.ID {
		return nil, nil, errors.New("direct transfers to system account are not allowed")
	}

	fromBalanceBefore := fromWallet.Balance
	fromBalanceAfter := fromBalanceBefore.Sub(amount)

	// Double-check sufficient funds within transaction
	if fromBalanceAfter.LessThan(decimal.Zero) {
		return nil, nil, errors.New("insufficient funds for transfer")
	}

	var outTransaction, inTransaction *models.Transaction

	err = uc.repos.DB.Transaction(func(tx *gorm.DB) error {
		outReference := fmt.Sprintf("%s-OUT", reference)
		fromBalanceBefore := fromWallet.Balance
		fromBalanceAfter := fromBalanceBefore.Sub(amount)

		if fromBalanceAfter.LessThan(decimal.Zero) {
			return errors.New("insufficient funds for transfer")
		}

		outTransaction = &models.Transaction{
			Reference:          outReference,
			WalletID:           fromWalletID,
			TransactionType:    models.TransactionTypeDebit,
			Amount:             amount,
			Metadata:           `{"source": "transfer"}`,
			BalanceBefore:      fromBalanceBefore,
			TransactionPurpose: "TRANSFER",
			BalanceAfter:       fromBalanceAfter,
			Description:        fmt.Sprintf("Transfer to wallet %d: %s", toWalletID, description),
			Status:             models.TransactionStatusCompleted,
		}

		if err := tx.Create(outTransaction).Error; err != nil {
			return fmt.Errorf("failed to create outgoing transaction: %w", err)
		}

		inReference := fmt.Sprintf("%s-IN", reference)
		toBalanceBefore := toWallet.Balance
		toBalanceAfter := toBalanceBefore.Add(amount)

		inTransaction = &models.Transaction{
			Reference:            inReference,
			WalletID:             toWalletID,
			TransactionType:      models.TransactionTypeCredit,
			TransactionPurpose:   "TRANSFER",
			Amount:               amount,
			BalanceBefore:        toBalanceBefore,
			Metadata:             `{"source": "transfer"}`,
			BalanceAfter:         toBalanceAfter,
			Description:          fmt.Sprintf("Transfer from wallet %d: %s", fromWalletID, description),
			Status:               models.TransactionStatusCompleted,
			RelatedTransactionID: &outTransaction.ID,
		}

		if err := tx.Create(inTransaction).Error; err != nil {
			return fmt.Errorf("failed to create incoming transaction: %w", err)
		}

		result := tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", fromWalletID, fromWallet.Version).
			Updates(map[string]interface{}{
				"balance": fromBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update source wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("source wallet version mismatch - concurrent modification detected")
		}

		result = tx.Model(&models.Wallet{}).Where("id = ? AND version = ?", toWalletID, toWallet.Version).
			Updates(map[string]interface{}{
				"balance": toBalanceAfter,
				"version": gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update destination wallet balance: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("destination wallet version mismatch - concurrent modification detected")
		}

		if err := tx.Model(outTransaction).Update("related_transaction_id", inTransaction.ID).Error; err != nil {
			return fmt.Errorf("failed to link outgoing transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// POST-TRANSACTION RECONCILIATION: Audit checks for both wallets
	go func() {
		uc.performPostTransactionReconciliation(fromWalletID)
		uc.performPostTransactionReconciliation(toWalletID)
	}()

	outTx, err := uc.repos.Transaction.GetByID(outTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load outgoing transaction: %w", err)
	}

	inTx, err := uc.repos.Transaction.GetByID(inTransaction.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load incoming transaction: %w", err)
	}

	return outTx, inTx, nil
}

func (uc *walletUseCase) GetWalletBalance(walletID uint) (decimal.Decimal, error) {
	wallet, err := uc.repos.Wallet.GetByID(walletID)
	if err != nil {
		return decimal.Zero, err
	}
	return wallet.Balance, nil
}

func (uc *walletUseCase) GetTransactionHistory(walletID uint, cursor *string, limit int) ([]models.Transaction, *string, error) {
	_, err := uc.repos.Wallet.GetByID(walletID)
	if err != nil {
		return nil, nil, errors.New("wallet not found")
	}

	var cursorTime *time.Time
	var cursorID *uint

	if cursor != nil && *cursor != "" {
		decodedCursor, err := uc.decodeCursor(*cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorTime = &decodedCursor.CreatedAt
		cursorID = &decodedCursor.ID
	}
	transactions, err := uc.repos.Transaction.GetByWalletIDWithCursor(walletID, cursorTime, cursorID, limit)
	if err != nil {
		return nil, nil, err
	}

	hasMore := len(transactions) > limit
	if hasMore {
		transactions = transactions[:limit] // Remove the extra transaction
	}

	var nextCursor *string

	if len(transactions) > 0 {
		if hasMore {
			lastTx := transactions[len(transactions)-1]
			nextCursorData := TransactionCursor{
				ID:        lastTx.ID,
				CreatedAt: lastTx.CreatedAt,
			}
			nextCursor, _ = uc.encodeCursor(nextCursorData)
		}
	}
	return transactions, nextCursor, nil
}

// encodeCursor encodes a cursor to a base64 string
func (uc *walletUseCase) encodeCursor(cursor TransactionCursor) (*string, error) {
	cursorJSON, err := json.Marshal(cursor)
	if err != nil {
		return nil, err
	}
	encodedCursor := base64.StdEncoding.EncodeToString(cursorJSON)
	return &encodedCursor, nil
}

// decodeCursor decodes a base64 cursor string
func (uc *walletUseCase) decodeCursor(cursor string) (*TransactionCursor, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var transactionCursor TransactionCursor
	err = json.Unmarshal(decodedBytes, &transactionCursor)
	if err != nil {
		return nil, err
	}

	return &transactionCursor, nil
}
