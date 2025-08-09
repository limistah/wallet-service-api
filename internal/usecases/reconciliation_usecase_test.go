package usecases

import (
	"fmt"
	"testing"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/shopspring/decimal"
)

// Helper function to set up test environment for reconciliation tests
func setupReconciliationTestEnvironment() *repositories.Repositories {
	userRepo := NewMockUserRepository()
	walletRepo := NewMockWalletRepository()
	transactionRepo := NewMockTransactionRepository()
	transactionTypeRepo := NewMockTransactionTypeRepository()
	reconciliationRepo := NewMockReconciliationRepository()

	// Create system user and wallet
	systemUser := &models.User{
		ID:    1,
		Email: models.SystemAccountEmail,
		Name:  "System Account",
	}
	userRepo.Create(systemUser)

	systemWallet := &models.Wallet{
		ID:       1,
		UserID:   systemUser.ID,
		Balance:  decimal.NewFromFloat(1000000.00), // Large system balance
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(systemWallet)

	repos := &repositories.Repositories{
		User:            userRepo,
		Wallet:          walletRepo,
		Transaction:     transactionRepo,
		TransactionType: transactionTypeRepo,
		Reconciliation:  reconciliationRepo,
		DB:              nil, // Skip DB for unit tests
	}

	return repos
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test Reconciliation functionality
func TestReconciliationUseCase_PerformWalletReconciliation(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	// Create test user and wallet
	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)
	transactionRepo := repos.Transaction.(*MockTransactionRepository)

	user := &models.User{
		ID:    20,
		Email: "reconcile@example.com",
		Name:  "Reconcile User",
	}
	userRepo.Create(user)

	wallet := &models.Wallet{
		ID:       20,
		UserID:   user.ID,
		Balance:  decimal.NewFromFloat(100.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(wallet)

	t.Run("should create reconciliation report for matching balance", func(t *testing.T) {
		// Create transactions that match the wallet balance
		tx1 := &models.Transaction{
			WalletID: 20,
			Amount:   decimal.NewFromFloat(100.00),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(tx1)

		report, err := reconciliationUC.PerformWalletReconciliation(20)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report == nil {
			t.Error("Expected reconciliation report, got nil")
		}

		if report != nil {
			if report.WalletID != 20 {
				t.Errorf("Expected wallet ID 20, got: %d", report.WalletID)
			}
			if !report.StoredBalance.Equal(decimal.NewFromFloat(100.00)) {
				t.Errorf("Expected stored balance 100.00, got: %v", report.StoredBalance)
			}
			if !report.CalculatedBalance.Equal(decimal.NewFromFloat(100.00)) {
				t.Errorf("Expected calculated balance 100.00, got: %v", report.CalculatedBalance)
			}
			if !report.Difference.IsZero() {
				t.Errorf("Expected difference 0, got: %v", report.Difference)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH, got: %v", report.Status)
			}
			if report.Notes != "Balance matches" {
				t.Errorf("Expected notes 'Balance matches', got: %s", report.Notes)
			}
		}
	})

	t.Run("should detect balance mismatch", func(t *testing.T) {
		// Create wallet with different balance than transactions
		mismatchUser := &models.User{
			ID:    21,
			Email: "mismatch@example.com",
			Name:  "Mismatch User",
		}
		userRepo.Create(mismatchUser)

		mismatchWallet := &models.Wallet{
			ID:       21,
			UserID:   mismatchUser.ID,
			Balance:  decimal.NewFromFloat(200.00), // Stored balance
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(mismatchWallet)

		// Create transactions that don't match the stored balance
		tx2 := &models.Transaction{
			WalletID: 21,
			Amount:   decimal.NewFromFloat(150.00), // Different from stored balance
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(tx2)

		report, err := reconciliationUC.PerformWalletReconciliation(21)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report == nil {
			t.Error("Expected reconciliation report, got nil")
		}

		if report != nil {
			if report.WalletID != 21 {
				t.Errorf("Expected wallet ID 21, got: %d", report.WalletID)
			}
			if !report.StoredBalance.Equal(decimal.NewFromFloat(200.00)) {
				t.Errorf("Expected stored balance 200.00, got: %v", report.StoredBalance)
			}
			if !report.CalculatedBalance.Equal(decimal.NewFromFloat(150.00)) {
				t.Errorf("Expected calculated balance 150.00, got: %v", report.CalculatedBalance)
			}
			expectedDifference := decimal.NewFromFloat(50.00)
			if !report.Difference.Equal(expectedDifference) {
				t.Errorf("Expected difference 50.00, got: %v", report.Difference)
			}
			if report.Status != models.ReconciliationStatusMismatch {
				t.Errorf("Expected status MISMATCH, got: %v", report.Status)
			}
			if !containsString(report.Notes, "Balance mismatch detected") {
				t.Errorf("Expected notes to contain 'Balance mismatch detected', got: %s", report.Notes)
			}
		}
	})

	t.Run("should handle nonexistent wallet", func(t *testing.T) {
		_, err := reconciliationUC.PerformWalletReconciliation(999)
		if err == nil {
			t.Error("Expected error for nonexistent wallet")
		}
	})
}

func TestReconciliationUseCase_PerformReconciliation(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	// Create test users and wallets
	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)
	transactionRepo := repos.Transaction.(*MockTransactionRepository)

	// Create first wallet with matching balance
	user1 := &models.User{
		ID:    22,
		Email: "bulk1@example.com",
		Name:  "Bulk User 1",
	}
	userRepo.Create(user1)

	wallet1 := &models.Wallet{
		ID:       22,
		UserID:   user1.ID,
		Balance:  decimal.NewFromFloat(300.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(wallet1)

	// Create matching transaction
	tx1 := &models.Transaction{
		WalletID: 22,
		Amount:   decimal.NewFromFloat(300.00),
		Status:   models.TransactionStatusCompleted,
	}
	transactionRepo.Create(tx1)

	// Create second wallet with mismatched balance
	user2 := &models.User{
		ID:    23,
		Email: "bulk2@example.com",
		Name:  "Bulk User 2",
	}
	userRepo.Create(user2)

	wallet2 := &models.Wallet{
		ID:       23,
		UserID:   user2.ID,
		Balance:  decimal.NewFromFloat(400.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(wallet2)

	// Create mismatched transaction
	tx2 := &models.Transaction{
		WalletID: 23,
		Amount:   decimal.NewFromFloat(350.00), // Different from wallet balance
		Status:   models.TransactionStatusCompleted,
	}
	transactionRepo.Create(tx2)

	t.Run("should perform bulk reconciliation for all wallets", func(t *testing.T) {
		reports, err := reconciliationUC.PerformReconciliation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(reports) == 0 {
			t.Error("Expected reconciliation reports, got empty slice")
		}

		// Should have reports for system wallet and our test wallets
		// Check if we have reports (exact count may vary due to system wallet)
		hasMatchReport := false
		hasMismatchReport := false

		for _, report := range reports {
			if report.Status == models.ReconciliationStatusMatch {
				hasMatchReport = true
			}
			if report.Status == models.ReconciliationStatusMismatch {
				hasMismatchReport = true
			}
		}

		if !hasMatchReport {
			t.Error("Expected at least one matching report")
		}
		if !hasMismatchReport {
			t.Error("Expected at least one mismatch report")
		}
	})
}

func TestReconciliationUseCase_GetReconciliationReports(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)
	reconciliationRepo := repos.Reconciliation.(*MockReconciliationRepository)

	// Create test reconciliation reports
	report1 := &models.ReconciliationReport{
		WalletID:          24,
		StoredBalance:     decimal.NewFromFloat(100.00),
		CalculatedBalance: decimal.NewFromFloat(100.00),
		Difference:        decimal.Zero,
		Status:            models.ReconciliationStatusMatch,
		Notes:             "Balance matches",
	}
	reconciliationRepo.Create(report1)

	report2 := &models.ReconciliationReport{
		WalletID:          25,
		StoredBalance:     decimal.NewFromFloat(200.00),
		CalculatedBalance: decimal.NewFromFloat(180.00),
		Difference:        decimal.NewFromFloat(20.00),
		Status:            models.ReconciliationStatusMismatch,
		Notes:             "Balance mismatch detected",
	}
	reconciliationRepo.Create(report2)

	t.Run("should get reconciliation reports with pagination", func(t *testing.T) {
		reports, err := reconciliationUC.GetReconciliationReports(1, 10)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(reports) == 0 {
			t.Error("Expected reconciliation reports, got empty slice")
		}

		// Verify we have our test reports
		foundReport1 := false
		foundReport2 := false

		for _, report := range reports {
			if report.WalletID == 24 && report.Status == models.ReconciliationStatusMatch {
				foundReport1 = true
			}
			if report.WalletID == 25 && report.Status == models.ReconciliationStatusMismatch {
				foundReport2 = true
			}
		}

		if !foundReport1 {
			t.Error("Expected to find report1 with MATCH status")
		}
		if !foundReport2 {
			t.Error("Expected to find report2 with MISMATCH status")
		}
	})
}

func TestReconciliationUseCase_GetMismatchReports(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)
	reconciliationRepo := repos.Reconciliation.(*MockReconciliationRepository)

	// Create test reconciliation reports - mix of match and mismatch
	matchReport := &models.ReconciliationReport{
		WalletID:          26,
		StoredBalance:     decimal.NewFromFloat(500.00),
		CalculatedBalance: decimal.NewFromFloat(500.00),
		Difference:        decimal.Zero,
		Status:            models.ReconciliationStatusMatch,
		Notes:             "Balance matches",
	}
	reconciliationRepo.Create(matchReport)

	mismatchReport1 := &models.ReconciliationReport{
		WalletID:          27,
		StoredBalance:     decimal.NewFromFloat(600.00),
		CalculatedBalance: decimal.NewFromFloat(580.00),
		Difference:        decimal.NewFromFloat(20.00),
		Status:            models.ReconciliationStatusMismatch,
		Notes:             "Balance mismatch detected",
	}
	reconciliationRepo.Create(mismatchReport1)

	mismatchReport2 := &models.ReconciliationReport{
		WalletID:          28,
		StoredBalance:     decimal.NewFromFloat(700.00),
		CalculatedBalance: decimal.NewFromFloat(750.00),
		Difference:        decimal.NewFromFloat(-50.00),
		Status:            models.ReconciliationStatusMismatch,
		Notes:             "Balance mismatch detected",
	}
	reconciliationRepo.Create(mismatchReport2)

	t.Run("should get only mismatch reports", func(t *testing.T) {
		reports, err := reconciliationUC.GetMismatchReports(1, 10)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(reports) == 0 {
			t.Error("Expected mismatch reports, got empty slice")
		}

		// Verify all reports are mismatches
		for _, report := range reports {
			if report.Status != models.ReconciliationStatusMismatch {
				t.Errorf("Expected only MISMATCH status, got: %v", report.Status)
			}
		}

		// Check for our specific mismatch reports
		foundMismatch1 := false
		foundMismatch2 := false
		foundMatch := false

		for _, report := range reports {
			if report.WalletID == 27 {
				foundMismatch1 = true
			}
			if report.WalletID == 28 {
				foundMismatch2 = true
			}
			if report.WalletID == 26 {
				foundMatch = true
			}
		}

		if !foundMismatch1 {
			t.Error("Expected to find mismatch report for wallet 27")
		}
		if !foundMismatch2 {
			t.Error("Expected to find mismatch report for wallet 28")
		}
		if foundMatch {
			t.Error("Should not find match report in mismatch results")
		}
	})
}

// Test reconciliation report model methods
func TestReconciliationReport_ModelMethods(t *testing.T) {
	t.Run("should correctly identify mismatch", func(t *testing.T) {
		report := &models.ReconciliationReport{
			Status: models.ReconciliationStatusMismatch,
		}

		if !report.HasMismatch() {
			t.Error("Expected HasMismatch to return true for MISMATCH status")
		}

		if !report.HasAnyIssue() {
			t.Error("Expected HasAnyIssue to return true for MISMATCH status")
		}

		if report.HasDoubleEntryError() {
			t.Error("Expected HasDoubleEntryError to return false for MISMATCH status")
		}
	})

	t.Run("should correctly identify match", func(t *testing.T) {
		report := &models.ReconciliationReport{
			Status: models.ReconciliationStatusMatch,
		}

		if report.HasMismatch() {
			t.Error("Expected HasMismatch to return false for MATCH status")
		}

		if report.HasAnyIssue() {
			t.Error("Expected HasAnyIssue to return false for MATCH status")
		}

		if report.HasDoubleEntryError() {
			t.Error("Expected HasDoubleEntryError to return false for MATCH status")
		}
	})

	t.Run("should correctly identify double entry error", func(t *testing.T) {
		report := &models.ReconciliationReport{
			Status: models.ReconciliationStatusDoubleEntryError,
		}

		if report.HasMismatch() {
			t.Error("Expected HasMismatch to return false for DOUBLE_ENTRY_ERROR status")
		}

		if !report.HasAnyIssue() {
			t.Error("Expected HasAnyIssue to return true for DOUBLE_ENTRY_ERROR status")
		}

		if !report.HasDoubleEntryError() {
			t.Error("Expected HasDoubleEntryError to return true for DOUBLE_ENTRY_ERROR status")
		}
	})
}

// Test edge cases and error scenarios
func TestReconciliationUseCase_EdgeCases(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should handle wallet with no transactions", func(t *testing.T) {
		// Create test user and wallet with no transactions
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)

		user := &models.User{
			ID:    29,
			Email: "notx@example.com",
			Name:  "No Transactions User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       29,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(100.00), // Non-zero balance but no transactions
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		report, err := reconciliationUC.PerformWalletReconciliation(29)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report == nil {
			t.Error("Expected reconciliation report, got nil")
		}

		if report != nil {
			// Should detect mismatch since wallet has balance but no transactions
			if report.Status != models.ReconciliationStatusMismatch {
				t.Errorf("Expected status MISMATCH, got: %v", report.Status)
			}
			if !report.CalculatedBalance.IsZero() {
				t.Errorf("Expected calculated balance 0 (no transactions), got: %v", report.CalculatedBalance)
			}
		}
	})

	t.Run("should handle wallet with pending transactions", func(t *testing.T) {
		// Create test user and wallet
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    30,
			Email: "pending@example.com",
			Name:  "Pending Transactions User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       30,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create pending transaction (should not be included in calculation)
		pendingTx := &models.Transaction{
			WalletID: 30,
			Amount:   decimal.NewFromFloat(50.00),
			Status:   models.TransactionStatusPending, // Not completed
		}
		transactionRepo.Create(pendingTx)

		// Create completed transaction
		completedTx := &models.Transaction{
			WalletID: 30,
			Amount:   decimal.NewFromFloat(100.00),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(completedTx)

		report, err := reconciliationUC.PerformWalletReconciliation(30)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report != nil {
			// Should only count completed transactions
			if !report.CalculatedBalance.Equal(decimal.NewFromFloat(100.00)) {
				t.Errorf("Expected calculated balance 100.00 (only completed tx), got: %v", report.CalculatedBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle wallet with failed transactions", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    31,
			Email: "failed@example.com",
			Name:  "Failed Transactions User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       31,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(150.00),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create failed transaction (should not be included in calculation)
		failedTx := &models.Transaction{
			WalletID: 31,
			Amount:   decimal.NewFromFloat(100.00),
			Status:   models.TransactionStatusFailed,
		}
		transactionRepo.Create(failedTx)

		// Create successful transaction
		successTx := &models.Transaction{
			WalletID: 31,
			Amount:   decimal.NewFromFloat(150.00),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(successTx)

		report, err := reconciliationUC.PerformWalletReconciliation(31)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report != nil {
			// Should only count completed transactions, ignoring failed ones
			if !report.CalculatedBalance.Equal(decimal.NewFromFloat(150.00)) {
				t.Errorf("Expected calculated balance 150.00 (only completed tx), got: %v", report.CalculatedBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle mixed debit and credit transactions", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    32,
			Email: "mixed@example.com",
			Name:  "Mixed Transactions User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       32,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(50.00), // Final expected balance
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create credit transaction (+200)
		creditTx := &models.Transaction{
			WalletID:        32,
			Amount:          decimal.NewFromFloat(200.00),
			TransactionType: models.TransactionTypeCredit,
			Status:          models.TransactionStatusCompleted,
		}
		transactionRepo.Create(creditTx)

		// Create debit transaction (-150)
		debitTx := &models.Transaction{
			WalletID:        32,
			Amount:          decimal.NewFromFloat(-150.00),
			TransactionType: models.TransactionTypeDebit,
			Status:          models.TransactionStatusCompleted,
		}
		transactionRepo.Create(debitTx)

		report, err := reconciliationUC.PerformWalletReconciliation(32)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report != nil {
			// Net balance should be 200 - 150 = 50
			expectedBalance := decimal.NewFromFloat(50.00)
			if !report.CalculatedBalance.Equal(expectedBalance) {
				t.Errorf("Expected calculated balance %v, got: %v", expectedBalance, report.CalculatedBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle suspended wallet", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)

		user := &models.User{
			ID:    33,
			Email: "suspended@example.com",
			Name:  "Suspended User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       33,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "USD",
			Status:   models.WalletStatusSuspended, // Suspended wallet
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Should still perform reconciliation even for suspended wallets
		report, err := reconciliationUC.PerformWalletReconciliation(33)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report == nil {
			t.Error("Expected reconciliation report even for suspended wallet")
		}
	})

	t.Run("should handle zero balances", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)

		user := &models.User{
			ID:    34,
			Email: "zero@example.com",
			Name:  "Zero Balance User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       34,
			UserID:   user.ID,
			Balance:  decimal.Zero, // Zero balance
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		report, err := reconciliationUC.PerformWalletReconciliation(34)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report != nil {
			if !report.StoredBalance.IsZero() {
				t.Errorf("Expected stored balance 0, got: %v", report.StoredBalance)
			}
			if !report.CalculatedBalance.IsZero() {
				t.Errorf("Expected calculated balance 0, got: %v", report.CalculatedBalance)
			}
			if !report.Difference.IsZero() {
				t.Errorf("Expected difference 0, got: %v", report.Difference)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH for zero balances, got: %v", report.Status)
			}
		}
	})
}

// Test system account reconciliation scenarios
func TestReconciliationUseCase_SystemAccountScenarios(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should reconcile system account", func(t *testing.T) {
		// System wallet should be ID 1 from setup
		report, err := reconciliationUC.PerformWalletReconciliation(1)
		if err != nil {
			t.Errorf("Expected no error reconciling system account, got: %v", err)
		}

		if report == nil {
			t.Error("Expected reconciliation report for system account")
		}

		if report != nil {
			if report.WalletID != 1 {
				t.Errorf("Expected system wallet ID 1, got: %d", report.WalletID)
			}
			// System account should have large initial balance
			if report.StoredBalance.LessThan(decimal.NewFromFloat(1000000.00)) {
				t.Errorf("Expected system account to have large balance, got: %v", report.StoredBalance)
			}
		}
	})
}

// Test boundary conditions and edge cases
func TestReconciliationUseCase_BoundaryConditions(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should handle large decimal values", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    35,
			Email: "large@example.com",
			Name:  "Large Values User",
		}
		userRepo.Create(user)

		// Large balance
		largeBalance := decimal.NewFromFloat(999999999.99)
		wallet := &models.Wallet{
			ID:       35,
			UserID:   user.ID,
			Balance:  largeBalance,
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Matching large transaction
		largeTx := &models.Transaction{
			WalletID: 35,
			Amount:   largeBalance,
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(largeTx)

		report, err := reconciliationUC.PerformWalletReconciliation(35)
		if err != nil {
			t.Errorf("Expected no error with large values, got: %v", err)
		}

		if report != nil {
			if !report.StoredBalance.Equal(largeBalance) {
				t.Errorf("Expected stored balance %v, got: %v", largeBalance, report.StoredBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH for large values, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle very small decimal differences", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    36,
			Email: "small@example.com",
			Name:  "Small Difference User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       36,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Transaction with tiny difference (100.01 vs 100.00)
		smallDiffTx := &models.Transaction{
			WalletID: 36,
			Amount:   decimal.NewFromFloat(100.01),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(smallDiffTx)

		report, err := reconciliationUC.PerformWalletReconciliation(36)
		if err != nil {
			t.Errorf("Expected no error with small differences, got: %v", err)
		}

		if report != nil {
			expectedDiff := decimal.NewFromFloat(-0.01)
			if !report.Difference.Equal(expectedDiff) {
				t.Errorf("Expected difference %v, got: %v", expectedDiff, report.Difference)
			}
			if report.Status != models.ReconciliationStatusMismatch {
				t.Errorf("Expected status MISMATCH for small difference, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle negative balances", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    37,
			Email: "negative@example.com",
			Name:  "Negative Balance User",
		}
		userRepo.Create(user)

		// Negative stored balance
		negativeBalance := decimal.NewFromFloat(-50.00)
		wallet := &models.Wallet{
			ID:       37,
			UserID:   user.ID,
			Balance:  negativeBalance,
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Matching negative transaction
		negativeTx := &models.Transaction{
			WalletID: 37,
			Amount:   negativeBalance,
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(negativeTx)

		report, err := reconciliationUC.PerformWalletReconciliation(37)
		if err != nil {
			t.Errorf("Expected no error with negative balances, got: %v", err)
		}

		if report != nil {
			if !report.StoredBalance.Equal(negativeBalance) {
				t.Errorf("Expected stored balance %v, got: %v", negativeBalance, report.StoredBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH for negative balances, got: %v", report.Status)
			}
		}
	})
}

// Test additional model methods and report functionality
func TestReconciliationReport_AdditionalMethods(t *testing.T) {
	t.Run("should correctly determine severity levels", func(t *testing.T) {
		matchReport := &models.ReconciliationReport{
			Status: models.ReconciliationStatusMatch,
		}
		if matchReport.GetSeverity() != "INFO" {
			t.Errorf("Expected INFO severity for MATCH, got: %s", matchReport.GetSeverity())
		}

		mismatchReport := &models.ReconciliationReport{
			Status: models.ReconciliationStatusMismatch,
		}
		if mismatchReport.GetSeverity() != "WARNING" {
			t.Errorf("Expected WARNING severity for MISMATCH, got: %s", mismatchReport.GetSeverity())
		}

		errorReport := &models.ReconciliationReport{
			Status: models.ReconciliationStatusDoubleEntryError,
		}
		if errorReport.GetSeverity() != "CRITICAL" {
			t.Errorf("Expected CRITICAL severity for DOUBLE_ENTRY_ERROR, got: %s", errorReport.GetSeverity())
		}
	})

	t.Run("should handle table name override", func(t *testing.T) {
		report := &models.ReconciliationReport{}
		if report.TableName() != "reconciliation_reports" {
			t.Errorf("Expected table name 'reconciliation_reports', got: %s", report.TableName())
		}
	})
}

// Test error handling and recovery scenarios
func TestReconciliationUseCase_ErrorHandling(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should handle repository errors gracefully", func(t *testing.T) {
		// Test with invalid wallet ID that doesn't exist
		_, err := reconciliationUC.PerformWalletReconciliation(99999)
		if err == nil {
			t.Error("Expected error for non-existent wallet")
		}
	})

	t.Run("should handle bulk reconciliation with some failures", func(t *testing.T) {
		// Bulk reconciliation should continue even if some wallets fail
		reports, err := reconciliationUC.PerformReconciliation()

		// Should not return error even if some individual reconciliations fail
		if err != nil {
			t.Errorf("Expected no error from bulk reconciliation, got: %v", err)
		}

		// Should still have some reports for wallets that succeeded
		if len(reports) == 0 {
			t.Error("Expected some reconciliation reports from bulk operation")
		}
	})

	t.Run("should validate pagination parameters", func(t *testing.T) {
		// Test with valid pagination
		_, err := reconciliationUC.GetReconciliationReports(1, 10)
		if err != nil {
			t.Errorf("Expected no error with valid pagination, got: %v", err)
		}

		// Test with edge case pagination (page 0 should be handled by repository)
		_, err = reconciliationUC.GetReconciliationReports(0, 10)
		if err != nil {
			t.Errorf("Expected repository to handle page 0, got: %v", err)
		}
	})
}

// Test performance and scalability scenarios
func TestReconciliationUseCase_Performance(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should handle multiple wallets efficiently", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		// Create multiple test wallets
		numWallets := 10
		for i := 50; i < 50+numWallets; i++ {
			user := &models.User{
				ID:    uint(i),
				Email: fmt.Sprintf("perf%d@example.com", i),
				Name:  fmt.Sprintf("Performance User %d", i),
			}
			userRepo.Create(user)

			wallet := &models.Wallet{
				ID:       uint(i),
				UserID:   uint(i),
				Balance:  decimal.NewFromFloat(float64(i * 100)),
				Currency: "USD",
				Status:   models.WalletStatusActive,
				Version:  0,
			}
			walletRepo.Create(wallet)

			// Create matching transaction
			tx := &models.Transaction{
				WalletID: uint(i),
				Amount:   decimal.NewFromFloat(float64(i * 100)),
				Status:   models.TransactionStatusCompleted,
			}
			transactionRepo.Create(tx)
		}

		// Perform bulk reconciliation
		reports, err := reconciliationUC.PerformReconciliation()
		if err != nil {
			t.Errorf("Expected no error with multiple wallets, got: %v", err)
		}

		// Should have reports for all wallets (including system wallet)
		if len(reports) < numWallets {
			t.Errorf("Expected at least %d reports, got: %d", numWallets, len(reports))
		}

		// Verify all reports are successful matches
		matchCount := 0
		for _, report := range reports {
			if report.Status == models.ReconciliationStatusMatch {
				matchCount++
			}
		}

		// At least our test wallets should match
		if matchCount < numWallets {
			t.Errorf("Expected at least %d matching reports, got: %d", numWallets, matchCount)
		}
	})

	t.Run("should handle wallets with many transactions", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    70,
			Email: "manytx@example.com",
			Name:  "Many Transactions User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       70,
			UserID:   70,
			Balance:  decimal.NewFromFloat(1000.00), // Final balance
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create many transactions that sum to 1000
		totalAmount := decimal.Zero
		numTransactions := 20
		amountPerTx := decimal.NewFromFloat(50.00) // 20 * 50 = 1000

		for i := 0; i < numTransactions; i++ {
			tx := &models.Transaction{
				WalletID: 70,
				Amount:   amountPerTx,
				Status:   models.TransactionStatusCompleted,
			}
			transactionRepo.Create(tx)
			totalAmount = totalAmount.Add(amountPerTx)
		}

		report, err := reconciliationUC.PerformWalletReconciliation(70)
		if err != nil {
			t.Errorf("Expected no error with many transactions, got: %v", err)
		}

		if report != nil {
			if !report.CalculatedBalance.Equal(totalAmount) {
				t.Errorf("Expected calculated balance %v, got: %v", totalAmount, report.CalculatedBalance)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH with many transactions, got: %v", report.Status)
			}
		}
	})
}

// Test concurrent reconciliation scenarios (simulated)
func TestReconciliationUseCase_Concurrency(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should handle sequential reconciliation requests", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)

		user := &models.User{
			ID:    80,
			Email: "concurrent@example.com",
			Name:  "Concurrent User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       80,
			UserID:   80,
			Balance:  decimal.NewFromFloat(500.00),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Perform multiple reconciliations of the same wallet
		numRuns := 5
		for i := 0; i < numRuns; i++ {
			report, err := reconciliationUC.PerformWalletReconciliation(80)
			if err != nil {
				t.Errorf("Run %d: Expected no error, got: %v", i+1, err)
			}
			if report == nil {
				t.Errorf("Run %d: Expected reconciliation report", i+1)
			}
		}
	})

	t.Run("should maintain data consistency across operations", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    81,
			Email: "consistency@example.com",
			Name:  "Consistency User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       81,
			UserID:   81,
			Balance:  decimal.NewFromFloat(300.00),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// First reconciliation (no transactions)
		report1, err := reconciliationUC.PerformWalletReconciliation(81)
		if err != nil {
			t.Errorf("First reconciliation failed: %v", err)
		}

		// Add transaction
		tx := &models.Transaction{
			WalletID: 81,
			Amount:   decimal.NewFromFloat(300.00),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(tx)

		// Second reconciliation (with transaction)
		report2, err := reconciliationUC.PerformWalletReconciliation(81)
		if err != nil {
			t.Errorf("Second reconciliation failed: %v", err)
		}

		// Compare results
		if report1 != nil && report2 != nil {
			if report1.Status == report2.Status {
				// Status changed from mismatch (no tx) to match (with tx)
				if report1.Status != models.ReconciliationStatusMismatch ||
					report2.Status != models.ReconciliationStatusMatch {
					t.Error("Expected status to change from MISMATCH to MATCH after adding transaction")
				}
			}
		}
	})
}

// Test advanced reconciliation scenarios
func TestReconciliationUseCase_AdvancedScenarios(t *testing.T) {
	repos := setupReconciliationTestEnvironment()
	reconciliationUC := NewReconciliationUseCase(repos)

	t.Run("should detect complex balance mismatches", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    90,
			Email: "complex@example.com",
			Name:  "Complex Mismatch User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       90,
			UserID:   90,
			Balance:  decimal.NewFromFloat(1000.00), // Stored balance
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create transactions that don't match stored balance
		transactions := []decimal.Decimal{
			decimal.NewFromFloat(200.00),  // +200
			decimal.NewFromFloat(300.00),  // +300
			decimal.NewFromFloat(-100.00), // -100
			decimal.NewFromFloat(150.00),  // +150
		}
		// Net: 200 + 300 - 100 + 150 = 550 (vs stored 1000)

		for _, amount := range transactions {
			tx := &models.Transaction{
				WalletID: 90,
				Amount:   amount,
				Status:   models.TransactionStatusCompleted,
			}
			transactionRepo.Create(tx)
		}

		report, err := reconciliationUC.PerformWalletReconciliation(90)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if report != nil {
			expectedCalculated := decimal.NewFromFloat(550.00)
			expectedDifference := decimal.NewFromFloat(450.00) // 1000 - 550

			if !report.CalculatedBalance.Equal(expectedCalculated) {
				t.Errorf("Expected calculated balance %v, got: %v", expectedCalculated, report.CalculatedBalance)
			}

			if !report.Difference.Equal(expectedDifference) {
				t.Errorf("Expected difference %v, got: %v", expectedDifference, report.Difference)
			}

			if report.Status != models.ReconciliationStatusMismatch {
				t.Errorf("Expected status MISMATCH, got: %v", report.Status)
			}

			if !containsString(report.Notes, "Balance mismatch detected") {
				t.Errorf("Expected notes to mention balance mismatch, got: %s", report.Notes)
			}
		}
	})

	t.Run("should handle precision edge cases", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    91,
			Email: "precision@example.com",
			Name:  "Precision User",
		}
		userRepo.Create(user)

		// Use precise decimal values
		preciseBalance := decimal.NewFromFloat(123.456789)
		wallet := &models.Wallet{
			ID:       91,
			UserID:   91,
			Balance:  preciseBalance,
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Create matching precise transaction
		preciseTx := &models.Transaction{
			WalletID: 91,
			Amount:   preciseBalance,
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(preciseTx)

		report, err := reconciliationUC.PerformWalletReconciliation(91)
		if err != nil {
			t.Errorf("Expected no error with precise values, got: %v", err)
		}

		if report != nil {
			if !report.Difference.IsZero() {
				t.Errorf("Expected zero difference with precise matching values, got: %v", report.Difference)
			}
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH with precise values, got: %v", report.Status)
			}
		}
	})

	t.Run("should handle currency-specific scenarios", func(t *testing.T) {
		userRepo := repos.User.(*MockUserRepository)
		walletRepo := repos.Wallet.(*MockWalletRepository)
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		user := &models.User{
			ID:    92,
			Email: "currency@example.com",
			Name:  "Currency User",
		}
		userRepo.Create(user)

		// Different currency wallet
		wallet := &models.Wallet{
			ID:       92,
			UserID:   92,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "EUR", // Different currency
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		tx := &models.Transaction{
			WalletID: 92,
			Amount:   decimal.NewFromFloat(100.00),
			Status:   models.TransactionStatusCompleted,
		}
		transactionRepo.Create(tx)

		report, err := reconciliationUC.PerformWalletReconciliation(92)
		if err != nil {
			t.Errorf("Expected no error with different currency, got: %v", err)
		}

		if report != nil {
			// Should still reconcile correctly regardless of currency
			if report.Status != models.ReconciliationStatusMatch {
				t.Errorf("Expected status MATCH regardless of currency, got: %v", report.Status)
			}
		}
	})
}
