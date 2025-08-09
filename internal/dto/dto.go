package dto

import (
	"time"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
)

// UserResponse represents user response data
type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Age       int       `json:"age" example:"30"`
} //@name UserResponse

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Age      int    `json:"age" example:"30"`
} //@name CreateUserRequest

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Name string `json:"name" example:"John Doe"`
	Age  int    `json:"age" example:"30"`
} //@name UpdateUserRequest

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
} //@name LoginRequest

// LoginResponse represents user login response
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
} //@name LoginResponse

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
} //@name ChangePasswordRequest

// WalletResponse represents wallet response data
type WalletResponse struct {
	ID       uint            `json:"id" example:"1"`
	UserID   uint            `json:"user_id" example:"1"`
	Balance  decimal.Decimal `json:"balance" example:"1000.50"`
	Currency string          `json:"currency" example:"USD"`
	Status   string          `json:"status" example:"ACTIVE"`
	Version  uint            `json:"version" example:"1"`
} //@name WalletResponse

// FundWalletRequest represents fund wallet request
type FundWalletRequest struct {
	Amount      decimal.Decimal `json:"amount" binding:"required" example:"100.50"`
	Reference   string          `json:"reference" binding:"required" example:"REF123456"`
	Description string          `json:"description" example:"Deposit from bank"`
} //@name FundWalletRequest

// WithdrawRequest represents withdraw request
type WithdrawRequest struct {
	Amount      decimal.Decimal `json:"amount" binding:"required" example:"50.25"`
	Reference   string          `json:"reference" binding:"required" example:"WTH123456"`
	Description string          `json:"description" example:"ATM withdrawal"`
} //@name WithdrawRequest

// TransferRequest represents transfer request
type TransferRequest struct {
	ToWalletID  uint            `json:"to_wallet_id" binding:"required" example:"2"`
	Amount      decimal.Decimal `json:"amount" binding:"required" example:"75.00"`
	Reference   string          `json:"reference" binding:"required" example:"TRF123456"`
	Description string          `json:"description" example:"Payment to friend"`
} //@name TransferRequest

// TransactionResponse represents transaction response data
type TransactionResponse struct {
	ID                 uint            `json:"id" example:"1"`
	CreatedAt          time.Time       `json:"created_at" example:"2023-01-01T00:00:00Z"`
	Reference          string          `json:"reference" example:"REF123456"`
	WalletID           uint            `json:"wallet_id" example:"1"`
	TransactionType    string          `json:"transaction_type" example:"CREDIT"`
	TransactionPurpose string          `json:"transaction_purpose" example:"WITHDRAWAL"`
	Amount             decimal.Decimal `json:"amount" example:"100.50"`
	BalanceBefore      decimal.Decimal `json:"balance_before" example:"900.00"`
	BalanceAfter       decimal.Decimal `json:"balance_after" example:"1000.50"`
	Description        string          `json:"description" example:"Deposit from bank"`
	Status             string          `json:"status" example:"COMPLETED"`
} //@name TransactionResponse

// TransactionHistoryResponse represents cursor-paginated transaction history
type TransactionHistoryResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   CursorPaginationMeta  `json:"pagination"`
} //@name TransactionHistoryResponse

// ReconciliationReportResponse represents reconciliation report data
type ReconciliationReportResponse struct {
	ID                uint            `json:"id" example:"1"`
	CreatedAt         time.Time       `json:"created_at" example:"2023-01-01T00:00:00Z"`
	WalletID          uint            `json:"wallet_id" example:"1"`
	StoredBalance     decimal.Decimal `json:"stored_balance" example:"1000.50"`
	CalculatedBalance decimal.Decimal `json:"calculated_balance" example:"1000.50"`
	Difference        decimal.Decimal `json:"difference" example:"0.00"`
	Status            string          `json:"status" example:"MATCH"`
	Notes             string          `json:"notes" example:"Balance matches"`
} //@name ReconciliationReportResponse

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page      int `json:"page" example:"1"`
	PageSize  int `json:"page_size" example:"20"`
	Total     int `json:"total" example:"100"`
	TotalPage int `json:"total_pages" example:"5"`
} //@name PaginationMeta

// CursorPaginationMeta represents cursor-based pagination metadata
type CursorPaginationMeta struct {
	PageSize    int     `json:"page_size" example:"20"`
	NextCursor  *string `json:"next_cursor,omitempty" example:"eyJpZCI6MTAwLCJjcmVhdGVkX2F0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ=="`
	HasNextPage bool    `json:"has_next_page" example:"true"`
} //@name CursorPaginationMeta

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation successful"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:""`
} //@name APIResponse

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Operation failed"`
	Error   string `json:"error" example:"Validation error"`
} //@name ErrorResponse

// BalanceResponse represents wallet balance response
type BalanceResponse struct {
	WalletID uint            `json:"wallet_id" example:"1"`
	Balance  decimal.Decimal `json:"balance" example:"1000.50"`
	Currency string          `json:"currency" example:"USD"`
} //@name BalanceResponse

// Helper functions to convert models to DTOs
func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
	}
}

func ToWalletResponse(wallet *models.Wallet) WalletResponse {
	return WalletResponse{
		ID:       wallet.ID,
		UserID:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
		Status:   string(wallet.Status),
		Version:  wallet.Version,
	}
}

func ToTransactionResponse(transaction *models.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:                 transaction.ID,
		CreatedAt:          transaction.CreatedAt,
		Reference:          transaction.Reference,
		WalletID:           transaction.WalletID,
		TransactionType:    string(transaction.TransactionType),
		TransactionPurpose: string(transaction.TransactionPurpose),
		Amount:             transaction.Amount,
		BalanceBefore:      transaction.BalanceBefore,
		BalanceAfter:       transaction.BalanceAfter,
		Description:        transaction.Description,
		Status:             string(transaction.Status),
	}
}

func ToReconciliationReportResponse(report *models.ReconciliationReport) ReconciliationReportResponse {
	return ReconciliationReportResponse{
		ID:                report.ID,
		CreatedAt:         report.CreatedAt,
		WalletID:          report.WalletID,
		StoredBalance:     report.StoredBalance,
		CalculatedBalance: report.CalculatedBalance,
		Difference:        report.Difference,
		Status:            string(report.Status),
		Notes:             report.Notes,
	}
}
