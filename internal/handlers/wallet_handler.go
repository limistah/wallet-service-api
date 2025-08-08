package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/limistah/wallet-service/internal/dto"
	"github.com/limistah/wallet-service/internal/middleware"
	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/usecases"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	walletUseCase usecases.WalletUseCase
}

func NewWalletHandler(walletUseCase usecases.WalletUseCase) *WalletHandler {
	return &WalletHandler{
		walletUseCase: walletUseCase,
	}
}

// getAuthenticatedUserWallet gets the wallet for the authenticated user
func (h *WalletHandler) getAuthenticatedUserWallet(c *gin.Context) (*models.Wallet, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	wallet, err := h.walletUseCase.GetWalletByUserID(userID)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetWallet godoc
//
//	@Summary		Get wallet by authenticated user
//	@Description	Retrieve wallet information for the authenticated user
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.APIResponse{data=dto.WalletResponse}
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/wallets/me [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	wallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Wallet retrieved successfully",
		Data:    dto.ToWalletResponse(wallet),
	})
}

// GetWalletBalance godoc
//
//	@Summary		Get wallet balance
//	@Description	Retrieve current balance of the authenticated user's wallet
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.APIResponse{data=dto.BalanceResponse}
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/wallets/me/balance [get]
func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	wallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Balance retrieved successfully",
		Data: dto.BalanceResponse{
			WalletID: wallet.ID,
			Balance:  wallet.Balance,
			Currency: wallet.Currency,
		},
	})
}

// FundWallet godoc
//
//	@Summary		Fund wallet
//	@Description	Add money to the authenticated user's wallet
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.FundWalletRequest	true	"Fund wallet request"
//	@Success		200		{object}	dto.APIResponse{data=dto.TransactionResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse	"Duplicate reference"
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/wallets/me/fund [post]
func (h *WalletHandler) FundWallet(c *gin.Context) {
	wallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	var req dto.FundWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Amount must be greater than zero",
			Error:   "invalid amount",
		})
		return
	}

	userTransaction, systemTransaction, err := h.walletUseCase.FundWallet(wallet.ID, req.Amount, req.Reference, req.Description)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "duplicate reference" {
			status = http.StatusConflict
		}
		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: "Failed to fund wallet",
			Error:   err.Error(),
		})
		return
	}

	// Return both transactions in the response
	response := map[string]interface{}{
		"user_transaction":   dto.ToTransactionResponse(userTransaction),
		"system_transaction": dto.ToTransactionResponse(systemTransaction),
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Wallet funded successfully",
		Data:    response,
	})
}

// WithdrawFunds godoc
//
//	@Summary		Withdraw funds
//	@Description	Withdraw money from the authenticated user's wallet
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.WithdrawRequest	true	"Withdraw request"
//	@Success		200		{object}	dto.APIResponse{data=dto.TransactionResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse	"Duplicate reference or insufficient funds"
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/wallets/me/withdraw [post]
func (h *WalletHandler) WithdrawFunds(c *gin.Context) {
	wallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Amount must be greater than zero",
			Error:   "invalid amount",
		})
		return
	}

	userTransaction, systemTransaction, err := h.walletUseCase.WithdrawFunds(wallet.ID, req.Amount, req.Reference, req.Description)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to withdraw funds"

		// Handle specific error types
		switch {
		case err.Error() == "insufficient funds":
			status = http.StatusConflict
			message = "Insufficient funds for withdrawal"
		case err.Error() == "duplicate reference":
			status = http.StatusConflict
			message = "Duplicate transaction reference"
		case strings.Contains(err.Error(), "balance mismatch detected"):
			status = http.StatusConflict
			message = "Wallet balance inconsistency detected. Please contact support."
		case strings.Contains(err.Error(), "reconciliation"):
			status = http.StatusServiceUnavailable
			message = "Wallet reconciliation in progress. Please try again later."
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Funds withdrawn successfully",
		Data: map[string]interface{}{
			"user_transaction":   dto.ToTransactionResponse(userTransaction),
			"system_transaction": dto.ToTransactionResponse(systemTransaction),
		},
	})
}

// TransferFunds godoc
//
//	@Summary		Transfer funds
//	@Description	Transfer money from authenticated user's wallet to another wallet
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.TransferRequest	true	"Transfer request"
//	@Success		200		{object}	dto.APIResponse{data=[]dto.TransactionResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse	"Duplicate reference or insufficient funds"
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/wallets/me/transfer [post]
func (h *WalletHandler) TransferFunds(c *gin.Context) {
	// Get the authenticated user's wallet as the source wallet
	fromWallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Source wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Validate amount
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Amount must be greater than zero",
			Error:   "invalid amount",
		})
		return
	}

	// Validate that source and destination are different
	if fromWallet.ID == req.ToWalletID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "Cannot transfer to the same wallet",
			Error:   "invalid transfer",
		})
		return
	}

	outTx, inTx, err := h.walletUseCase.TransferFunds(fromWallet.ID, req.ToWalletID, req.Amount, req.Reference, req.Description)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to transfer funds"

		// Handle specific error types
		switch {
		case err.Error() == "insufficient funds":
			status = http.StatusConflict
			message = "Insufficient funds for transfer"
		case err.Error() == "duplicate reference":
			status = http.StatusConflict
			message = "Duplicate transaction reference"
		case strings.Contains(err.Error(), "balance mismatch detected"):
			status = http.StatusConflict
			message = "Wallet balance inconsistency detected. Please contact support."
		case strings.Contains(err.Error(), "reconciliation"):
			status = http.StatusServiceUnavailable
			message = "Wallet reconciliation in progress. Please try again later."
		case strings.Contains(err.Error(), "source wallet"):
			status = http.StatusNotFound
			message = "Source wallet not found or access denied"
		case strings.Contains(err.Error(), "destination wallet"):
			status = http.StatusNotFound
			message = "Destination wallet not found or inactive"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Funds transferred successfully",
		Data: []dto.TransactionResponse{
			dto.ToTransactionResponse(outTx),
			dto.ToTransactionResponse(inTx),
		},
	})
}

// GetTransactionHistory godoc
//
//	@Summary		Get transaction history
//	@Description	Retrieve paginated transaction history for the authenticated user's wallet
//	@Tags			wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Page number"	default(1)
//	@Param			limit	query		int	false	"Page size"		default(20)
//	@Success		200		{object}	dto.APIResponse{data=dto.TransactionHistoryResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/wallets/me/transactions [get]
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
	wallet, err := h.getAuthenticatedUserWallet(c)
	if err != nil {
		status := http.StatusNotFound
		message := "Wallet not found"

		if err.Error() == "user not authenticated" {
			status = http.StatusUnauthorized
			message = "User not authenticated"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if ps := c.Query("limit"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	transactions, err := h.walletUseCase.GetTransactionHistory(wallet.ID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve transaction history",
			Error:   err.Error(),
		})
		return
	}

	// Convert to DTOs
	transactionResponses := make([]dto.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		transactionResponses[i] = dto.ToTransactionResponse(&tx)
	}

	response := dto.TransactionHistoryResponse{
		Transactions: transactionResponses,
		Pagination: dto.PaginationMeta{
			Page:      page,
			PageSize:  pageSize,
			Total:     len(transactionResponses), // This should be total count from DB in real implementation
			TotalPage: 1,                         // Calculate based on total count
		},
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Transaction history retrieved successfully",
		Data:    response,
	})
}
