package usecases

import (
	"errors"
	"testing"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockUserRepository implements UserRepository interface for testing
type MockUserRepository struct {
	users  map[uint]*models.User
	emails map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint]*models.User),
		emails: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if user.ID == 0 {
		user.ID = uint(len(m.users) + 1)
	}
	m.users[user.ID] = user
	m.emails[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if user, ok := m.emails[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) Update(user *models.User) error {
	m.users[user.ID] = user
	m.emails[user.Email] = user
	return nil
}

func (m *MockUserRepository) Delete(id uint) error {
	if user, ok := m.users[id]; ok {
		delete(m.emails, user.Email)
		delete(m.users, id)
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (m *MockUserRepository) List(offset, limit int) ([]models.User, error) {
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

// MockWalletRepository implements WalletRepository interface for testing
type MockWalletRepository struct {
	wallets     map[uint]*models.Wallet
	userWallets map[uint]*models.Wallet
}

func NewMockWalletRepository() *MockWalletRepository {
	return &MockWalletRepository{
		wallets:     make(map[uint]*models.Wallet),
		userWallets: make(map[uint]*models.Wallet),
	}
}

func (m *MockWalletRepository) Create(wallet *models.Wallet) error {
	if wallet.ID == 0 {
		wallet.ID = uint(len(m.wallets) + 1)
	}
	m.wallets[wallet.ID] = wallet
	m.userWallets[wallet.UserID] = wallet
	return nil
}

func (m *MockWalletRepository) GetByID(id uint) (*models.Wallet, error) {
	if wallet, ok := m.wallets[id]; ok {
		return wallet, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockWalletRepository) GetByUserID(userID uint) (*models.Wallet, error) {
	if wallet, ok := m.userWallets[userID]; ok {
		return wallet, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockWalletRepository) Update(wallet *models.Wallet) error {
	m.wallets[wallet.ID] = wallet
	m.userWallets[wallet.UserID] = wallet
	return nil
}

func (m *MockWalletRepository) UpdateBalance(walletID uint, newBalance decimal.Decimal, version uint) error {
	if wallet, ok := m.wallets[walletID]; ok {
		if wallet.Version != version {
			return errors.New("version mismatch")
		}
		wallet.Balance = newBalance
		wallet.Version++
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (m *MockWalletRepository) List(offset, limit int) ([]models.Wallet, error) {
	wallets := make([]models.Wallet, 0, len(m.wallets))
	for _, wallet := range m.wallets {
		wallets = append(wallets, *wallet)
	}
	return wallets, nil
}

func (m *MockWalletRepository) GetAllForReconciliation() ([]models.Wallet, error) {
	return m.List(0, 100)
}

// MockTransactionRepository implements TransactionRepository interface for testing
type MockTransactionRepository struct {
	transactions map[uint]*models.Transaction
	references   map[string]*models.Transaction
	idCounter    uint
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[uint]*models.Transaction),
		references:   make(map[string]*models.Transaction),
		idCounter:    0,
	}
}

func (m *MockTransactionRepository) Create(transaction *models.Transaction) error {
	m.idCounter++
	transaction.ID = m.idCounter
	m.transactions[transaction.ID] = transaction
	m.references[transaction.Reference] = transaction
	return nil
}

func (m *MockTransactionRepository) GetByID(id uint) (*models.Transaction, error) {
	if transaction, ok := m.transactions[id]; ok {
		return transaction, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockTransactionRepository) GetByReference(reference string) (*models.Transaction, error) {
	if transaction, ok := m.references[reference]; ok {
		return transaction, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockTransactionRepository) GetByWalletID(walletID uint, offset, limit int) ([]models.Transaction, error) {
	transactions := make([]models.Transaction, 0)
	for _, transaction := range m.transactions {
		if transaction.WalletID == walletID {
			transactions = append(transactions, *transaction)
		}
	}
	return transactions, nil
}

func (m *MockTransactionRepository) Update(transaction *models.Transaction) error {
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) CalculateBalance(walletID uint) (decimal.Decimal, error) {
	balance := decimal.Zero
	for _, transaction := range m.transactions {
		if transaction.WalletID == walletID && transaction.Status == models.TransactionStatusCompleted {
			balance = balance.Add(transaction.Amount)
		}
	}
	return balance, nil
}

func (m *MockTransactionRepository) List(offset, limit int) ([]models.Transaction, error) {
	transactions := make([]models.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, *transaction)
	}
	return transactions, nil
}

// MockTransactionTypeRepository implements TransactionTypeRepository interface for testing
type MockTransactionTypeRepository struct {
	types map[string]*models.TransactionType
}

func NewMockTransactionTypeRepository() *MockTransactionTypeRepository {
	repo := &MockTransactionTypeRepository{
		types: make(map[string]*models.TransactionType),
	}
	// Pre-populate with default types
	repo.types[models.TransactionTypeCredit] = &models.TransactionType{
		ID:   1,
		Name: models.TransactionTypeCredit,
	}
	repo.types[models.TransactionTypeDebit] = &models.TransactionType{
		ID:   2,
		Name: models.TransactionTypeDebit,
	}
	return repo
}

// MockReconciliationRepository implements ReconciliationRepository interface for testing
type MockReconciliationRepository struct {
	reports map[uint]*models.ReconciliationReport
}

func NewMockReconciliationRepository() *MockReconciliationRepository {
	return &MockReconciliationRepository{
		reports: make(map[uint]*models.ReconciliationReport),
	}
}

func (m *MockReconciliationRepository) Create(report *models.ReconciliationReport) error {
	if report.ID == 0 {
		report.ID = uint(len(m.reports) + 1)
	}
	m.reports[report.ID] = report
	return nil
}

func (m *MockReconciliationRepository) GetByID(id uint) (*models.ReconciliationReport, error) {
	if report, ok := m.reports[id]; ok {
		return report, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockReconciliationRepository) GetByWalletID(walletID uint) ([]models.ReconciliationReport, error) {
	reports := make([]models.ReconciliationReport, 0)
	for _, report := range m.reports {
		if report.WalletID == walletID {
			reports = append(reports, *report)
		}
	}
	return reports, nil
}

func (m *MockReconciliationRepository) List(offset, limit int) ([]models.ReconciliationReport, error) {
	reports := make([]models.ReconciliationReport, 0, len(m.reports))
	for _, report := range m.reports {
		reports = append(reports, *report)
	}
	return reports, nil
}

func (m *MockReconciliationRepository) GetMismatches(offset, limit int) ([]models.ReconciliationReport, error) {
	reports := make([]models.ReconciliationReport, 0)
	for _, report := range m.reports {
		if report.Status == models.ReconciliationStatusMismatch {
			reports = append(reports, *report)
		}
	}
	return reports, nil
}

// MockReconciliationUseCase implements ReconciliationUseCase interface for testing
type MockReconciliationUseCase struct{}

func (m *MockReconciliationUseCase) PerformReconciliation() ([]models.ReconciliationReport, error) {
	return []models.ReconciliationReport{}, nil
}

func (m *MockReconciliationUseCase) PerformWalletReconciliation(walletID uint) (*models.ReconciliationReport, error) {
	// Return a successful reconciliation report
	return &models.ReconciliationReport{
		WalletID:          walletID,
		Status:            models.ReconciliationStatusMatch,
		StoredBalance:     decimal.NewFromFloat(100.00),
		CalculatedBalance: decimal.NewFromFloat(100.00),
		Difference:        decimal.Zero,
	}, nil
}

func (m *MockReconciliationUseCase) GetReconciliationReports(page, pageSize int) ([]models.ReconciliationReport, error) {
	return []models.ReconciliationReport{}, nil
}

func (m *MockReconciliationUseCase) GetMismatchReports(page, pageSize int) ([]models.ReconciliationReport, error) {
	return []models.ReconciliationReport{}, nil
}

func (m *MockTransactionTypeRepository) GetByName(name string) (*models.TransactionType, error) {
	if transactionType, ok := m.types[name]; ok {
		return transactionType, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockTransactionTypeRepository) List() ([]models.TransactionType, error) {
	types := make([]models.TransactionType, 0, len(m.types))
	for _, transactionType := range m.types {
		types = append(types, *transactionType)
	}
	return types, nil
}

func (m *MockTransactionTypeRepository) Create(transactionType *models.TransactionType) error {
	if transactionType.ID == 0 {
		transactionType.ID = uint(len(m.types) + 1)
	}
	m.types[transactionType.Name] = transactionType
	return nil
}

// Helper function to create test repositories and data
func setupTestEnvironment() (*repositories.Repositories, *MockReconciliationUseCase) {
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

	reconciliationUC := &MockReconciliationUseCase{}
	return repos, reconciliationUC
}

// Test Fund Wallet functionality
func TestWalletUseCase_FundWallet(t *testing.T) {
	repos, reconciliationUC := setupTestEnvironment()
	walletUC := NewWalletUseCase(repos, reconciliationUC)

	// Create test user and wallet
	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)

	user := &models.User{
		ID:    2,
		Email: "testuser@example.com",
		Name:  "Test User",
	}
	userRepo.Create(user)

	wallet := &models.Wallet{
		ID:       2,
		UserID:   user.ID,
		Balance:  decimal.NewFromFloat(100.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(wallet)

	t.Run("should reject zero amount", func(t *testing.T) {
		_, _, err := walletUC.FundWallet(2, decimal.Zero, "REF001", "Test funding")
		if err == nil {
			t.Error("Expected error for zero amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject negative amount", func(t *testing.T) {
		_, _, err := walletUC.FundWallet(2, decimal.NewFromFloat(-50.00), "REF002", "Test funding")
		if err == nil {
			t.Error("Expected error for negative amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject funding nonexistent wallet", func(t *testing.T) {
		_, _, err := walletUC.FundWallet(999, decimal.NewFromFloat(50.00), "REF003", "Test funding")
		if err == nil {
			t.Error("Expected error for nonexistent wallet")
		}
		// The error might come from reconciliation check or wallet lookup
		expectedErrors := []string{"wallet not found", "pre-transaction reconciliation failed: wallet not found"}
		found := false
		for _, expectedErr := range expectedErrors {
			if err.Error() == expectedErr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected wallet not found error, got: %v", err)
		}
	})

	t.Run("should reject duplicate reference", func(t *testing.T) {
		transactionRepo := repos.Transaction.(*MockTransactionRepository)

		// Create existing transaction with same reference
		existingTx := &models.Transaction{
			Reference: "DUPLICATE_REF",
			WalletID:  2,
			Amount:    decimal.NewFromFloat(25.00),
		}
		transactionRepo.Create(existingTx)

		_, _, err := walletUC.FundWallet(2, decimal.NewFromFloat(50.00), "DUPLICATE_REF", "Test funding")
		if err == nil {
			t.Error("Expected error for duplicate reference")
		}
		if err.Error() != "duplicate reference" {
			t.Errorf("Expected 'duplicate reference', got: %v", err)
		}
	})

	t.Run("should validate wallet status", func(t *testing.T) {
		// Create inactive wallet
		inactiveUser := &models.User{
			ID:    3,
			Email: "inactive@example.com",
			Name:  "Inactive User",
		}
		userRepo.Create(inactiveUser)

		inactiveWallet := &models.Wallet{
			ID:       3,
			UserID:   inactiveUser.ID,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "USD",
			Status:   models.WalletStatusSuspended, // Inactive status
			Version:  0,
		}
		walletRepo.Create(inactiveWallet)

		_, _, err := walletUC.FundWallet(3, decimal.NewFromFloat(50.00), "REF004", "Test funding")
		if err == nil {
			t.Error("Expected error for inactive wallet")
		}
		if err.Error() != "wallet is not active" {
			t.Errorf("Expected 'wallet is not active', got: %v", err)
		}
	})
}

// Test Withdraw Funds functionality
func TestWalletUseCase_WithdrawFunds(t *testing.T) {
	repos, reconciliationUC := setupTestEnvironment()
	walletUC := NewWalletUseCase(repos, reconciliationUC)

	// Create test user and wallet
	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)

	user := &models.User{
		ID:    4,
		Email: "withdraw@example.com",
		Name:  "Withdraw User",
	}
	userRepo.Create(user)

	wallet := &models.Wallet{
		ID:       4,
		UserID:   user.ID,
		Balance:  decimal.NewFromFloat(100.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(wallet)

	t.Run("should reject zero amount", func(t *testing.T) {
		_, _, err := walletUC.WithdrawFunds(4, decimal.Zero, "WD001", "Test withdrawal")
		if err == nil {
			t.Error("Expected error for zero amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject negative amount", func(t *testing.T) {
		_, _, err := walletUC.WithdrawFunds(4, decimal.NewFromFloat(-50.00), "WD002", "Test withdrawal")
		if err == nil {
			t.Error("Expected error for negative amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject withdrawal from nonexistent wallet", func(t *testing.T) {
		_, _, err := walletUC.WithdrawFunds(999, decimal.NewFromFloat(50.00), "WD003", "Test withdrawal")
		if err == nil {
			t.Error("Expected error for nonexistent wallet")
		}
		// Error could be from reconciliation or wallet lookup
		expectedErrors := []string{"wallet not found", "pre-transaction reconciliation failed: wallet not found"}
		found := false
		for _, expectedErr := range expectedErrors {
			if err.Error() == expectedErr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected wallet not found error, got: %v", err)
		}
	})

	t.Run("should reject withdrawal exceeding balance", func(t *testing.T) {
		_, _, err := walletUC.WithdrawFunds(4, decimal.NewFromFloat(200.00), "WD004", "Test withdrawal")
		if err == nil {
			t.Error("Expected error for insufficient funds")
		}
		// Should contain insufficient funds message
		if !contains(err.Error(), "insufficient funds") {
			t.Errorf("Expected 'insufficient funds' error, got: %v", err)
		}
	})

	t.Run("should reject withdrawal from inactive wallet", func(t *testing.T) {
		// Create inactive wallet
		inactiveUser := &models.User{
			ID:    5,
			Email: "inactive_wd@example.com",
			Name:  "Inactive Withdraw User",
		}
		userRepo.Create(inactiveUser)

		inactiveWallet := &models.Wallet{
			ID:       5,
			UserID:   inactiveUser.ID,
			Balance:  decimal.NewFromFloat(100.00),
			Currency: "USD",
			Status:   models.WalletStatusSuspended,
			Version:  0,
		}
		walletRepo.Create(inactiveWallet)

		_, _, err := walletUC.WithdrawFunds(5, decimal.NewFromFloat(50.00), "WD005", "Test withdrawal")
		if err == nil {
			t.Error("Expected error for inactive wallet")
		}
		if err.Error() != "wallet is not active" {
			t.Errorf("Expected 'wallet is not active', got: %v", err)
		}
	})
}

// Test Transfer Funds functionality
func TestWalletUseCase_TransferFunds(t *testing.T) {
	repos, reconciliationUC := setupTestEnvironment()
	walletUC := NewWalletUseCase(repos, reconciliationUC)

	// Create test users and wallets
	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)

	// Source user and wallet
	sourceUser := &models.User{
		ID:    6,
		Email: "source@example.com",
		Name:  "Source User",
	}
	userRepo.Create(sourceUser)

	sourceWallet := &models.Wallet{
		ID:       6,
		UserID:   sourceUser.ID,
		Balance:  decimal.NewFromFloat(200.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(sourceWallet)

	// Destination user and wallet
	destUser := &models.User{
		ID:    7,
		Email: "dest@example.com",
		Name:  "Destination User",
	}
	userRepo.Create(destUser)

	destWallet := &models.Wallet{
		ID:       7,
		UserID:   destUser.ID,
		Balance:  decimal.NewFromFloat(50.00),
		Currency: "USD",
		Status:   models.WalletStatusActive,
		Version:  0,
	}
	walletRepo.Create(destWallet)

	t.Run("should reject transfer to same wallet", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 6, decimal.NewFromFloat(50.00), "TR001", "Self transfer")
		if err == nil {
			t.Error("Expected error for transfer to same wallet")
		}
		if err.Error() != "cannot transfer to the same wallet" {
			t.Errorf("Expected 'cannot transfer to the same wallet', got: %v", err)
		}
	})

	t.Run("should reject zero amount", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 7, decimal.Zero, "TR002", "Zero transfer")
		if err == nil {
			t.Error("Expected error for zero amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject negative amount", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 7, decimal.NewFromFloat(-50.00), "TR003", "Negative transfer")
		if err == nil {
			t.Error("Expected error for negative amount")
		}
		if err.Error() != "amount must be greater than zero" {
			t.Errorf("Expected 'amount must be greater than zero', got: %v", err)
		}
	})

	t.Run("should reject transfer to nonexistent destination", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 999, decimal.NewFromFloat(50.00), "TR004", "Transfer to nowhere")
		if err == nil {
			t.Error("Expected error for nonexistent destination")
		}
		if err.Error() != "receiving wallet not found" {
			t.Errorf("Expected 'receiving wallet not found', got: %v", err)
		}
	})

	t.Run("should reject transfer from nonexistent source", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(999, 7, decimal.NewFromFloat(50.00), "TR005", "Transfer from nowhere")
		if err == nil {
			t.Error("Expected error for nonexistent source")
		}
		// Error could be from reconciliation check
		if !contains(err.Error(), "wallet") && !contains(err.Error(), "reconciliation") {
			t.Errorf("Expected wallet-related error, got: %v", err)
		}
	})

	t.Run("should reject transfer exceeding source balance", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 7, decimal.NewFromFloat(500.00), "TR006", "Excessive transfer")
		if err == nil {
			t.Error("Expected error for insufficient funds")
		}
		// Should fail at some validation step
		if !contains(err.Error(), "insufficient funds") && !contains(err.Error(), "reconciliation") {
			t.Errorf("Expected insufficient funds or reconciliation error, got: %v", err)
		}
	})

	t.Run("should reject transfer to inactive destination", func(t *testing.T) {
		// Create inactive destination wallet
		inactiveDestUser := &models.User{
			ID:    8,
			Email: "inactive_dest@example.com",
			Name:  "Inactive Dest User",
		}
		userRepo.Create(inactiveDestUser)

		inactiveDestWallet := &models.Wallet{
			ID:       8,
			UserID:   inactiveDestUser.ID,
			Balance:  decimal.NewFromFloat(0.00),
			Currency: "USD",
			Status:   models.WalletStatusSuspended,
			Version:  0,
		}
		walletRepo.Create(inactiveDestWallet)

		_, _, err := walletUC.TransferFunds(6, 8, decimal.NewFromFloat(50.00), "TR007", "Transfer to inactive")
		if err == nil {
			t.Error("Expected error for inactive destination wallet")
		}
		if err.Error() != "destination wallet is not active" {
			t.Errorf("Expected 'destination wallet is not active', got: %v", err)
		}
	})

	t.Run("should prevent transfer to system wallet", func(t *testing.T) {
		_, _, err := walletUC.TransferFunds(6, 1, decimal.NewFromFloat(50.00), "TR008", "Transfer to system")
		if err == nil {
			t.Error("Expected error for transfer to system wallet")
		}
		if err.Error() != "direct transfers to system account are not allowed" {
			t.Errorf("Expected 'direct transfers to system account are not allowed', got: %v", err)
		}
	})
}

// Test additional business logic methods
func TestWalletUseCase_BusinessLogic(t *testing.T) {
	repos, reconciliationUC := setupTestEnvironment()
	walletUC := NewWalletUseCase(repos, reconciliationUC)

	userRepo := repos.User.(*MockUserRepository)
	walletRepo := repos.Wallet.(*MockWalletRepository)

	t.Run("should create wallet successfully", func(t *testing.T) {
		user := &models.User{
			ID:    10,
			Email: "newwallet@example.com",
			Name:  "New Wallet User",
		}
		userRepo.Create(user)

		wallet, err := walletUC.CreateWallet(10, "USD")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if wallet == nil {
			t.Error("Expected wallet to be created")
		}

		if wallet != nil {
			if wallet.UserID != 10 {
				t.Errorf("Expected wallet user ID 10, got: %d", wallet.UserID)
			}
			if wallet.Currency != "USD" {
				t.Errorf("Expected wallet currency USD, got: %s", wallet.Currency)
			}
			if !wallet.Balance.Equal(decimal.Zero) {
				t.Errorf("Expected wallet balance 0, got: %v", wallet.Balance)
			}
			if wallet.Status != models.WalletStatusActive {
				t.Errorf("Expected wallet status ACTIVE, got: %v", wallet.Status)
			}
		}
	})

	t.Run("should prevent duplicate wallet creation", func(t *testing.T) {
		user := &models.User{
			ID:    11,
			Email: "duplicate@example.com",
			Name:  "Duplicate User",
		}
		userRepo.Create(user)

		// Create first wallet
		wallet := &models.Wallet{
			ID:       11,
			UserID:   user.ID,
			Balance:  decimal.Zero,
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		// Try to create second wallet
		_, err := walletUC.CreateWallet(11, "USD")
		if err == nil {
			t.Error("Expected error for duplicate wallet creation")
		}
		if err.Error() != "user already has a wallet" {
			t.Errorf("Expected 'user already has a wallet', got: %v", err)
		}
	})

	t.Run("should get wallet by ID", func(t *testing.T) {
		user := &models.User{
			ID:    12,
			Email: "getwallet@example.com",
			Name:  "Get Wallet User",
		}
		userRepo.Create(user)

		wallet := &models.Wallet{
			ID:       12,
			UserID:   user.ID,
			Balance:  decimal.NewFromFloat(150.75),
			Currency: "USD",
			Status:   models.WalletStatusActive,
			Version:  0,
		}
		walletRepo.Create(wallet)

		retrievedWallet, err := walletUC.GetWallet(12)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if retrievedWallet == nil {
			t.Error("Expected wallet to be retrieved")
		}

		if retrievedWallet != nil {
			if retrievedWallet.ID != 12 {
				t.Errorf("Expected wallet ID 12, got: %d", retrievedWallet.ID)
			}
			if !retrievedWallet.Balance.Equal(decimal.NewFromFloat(150.75)) {
				t.Errorf("Expected wallet balance 150.75, got: %v", retrievedWallet.Balance)
			}
		}
	})

	t.Run("should get wallet balance", func(t *testing.T) {
		balance, err := walletUC.GetWalletBalance(12)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		expectedBalance := decimal.NewFromFloat(150.75)
		if !balance.Equal(expectedBalance) {
			t.Errorf("Expected balance %v, got: %v", expectedBalance, balance)
		}
	})

	t.Run("should get wallet by user ID", func(t *testing.T) {
		retrievedWallet, err := walletUC.GetWalletByUserID(12)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if retrievedWallet == nil {
			t.Error("Expected wallet to be retrieved")
		}

		if retrievedWallet != nil {
			if retrievedWallet.UserID != 12 {
				t.Errorf("Expected wallet user ID 12, got: %d", retrievedWallet.UserID)
			}
		}
	})

	t.Run("should handle nonexistent wallet", func(t *testing.T) {
		_, err := walletUC.GetWallet(999)
		if err == nil {
			t.Error("Expected error for nonexistent wallet")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
