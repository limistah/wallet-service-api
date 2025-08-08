package usecases

import (
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
}
