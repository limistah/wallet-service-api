package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/limistah/wallet-service/internal/dto"
	"github.com/limistah/wallet-service/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWalletUseCase is a mock implementation of WalletUseCase for testing
type MockWalletUseCase struct {
	mock.Mock
}

func (m *MockWalletUseCase) CreateWallet(userID uint, currency string) (*models.Wallet, error) {
	args := m.Called(userID, currency)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletUseCase) GetWallet(id uint) (*models.Wallet, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletUseCase) GetWalletByUserID(userID uint) (*models.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletUseCase) FundWallet(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	args := m.Called(walletID, amount, reference, description)
	return args.Get(0).(*models.Transaction), args.Get(1).(*models.Transaction), args.Error(2)
}

func (m *MockWalletUseCase) WithdrawFunds(walletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	args := m.Called(walletID, amount, reference, description)
	return args.Get(0).(*models.Transaction), args.Get(1).(*models.Transaction), args.Error(2)
}

func (m *MockWalletUseCase) TransferFunds(fromWalletID, toWalletID uint, amount decimal.Decimal, reference, description string) (*models.Transaction, *models.Transaction, error) {
	args := m.Called(fromWalletID, toWalletID, amount, reference, description)
	return args.Get(0).(*models.Transaction), args.Get(1).(*models.Transaction), args.Error(2)
}

func (m *MockWalletUseCase) GetWalletBalance(walletID uint) (decimal.Decimal, error) {
	args := m.Called(walletID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockWalletUseCase) GetTransactionHistory(walletID uint, cursor *string, limit int) ([]models.Transaction, *string, error) {
	args := m.Called(walletID, cursor, limit)
	return args.Get(0).([]models.Transaction), args.Get(1).(*string), args.Error(2)
}

func createTestCursor(id uint, createdAt time.Time) string {
	type TransactionCursor struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
	}

	cursor := TransactionCursor{
		ID:        id,
		CreatedAt: createdAt,
	}

	cursorJSON, _ := json.Marshal(cursor)
	return base64.StdEncoding.EncodeToString(cursorJSON)
}

func TestWalletHandler_GetTransactionHistoryWithCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockWalletUseCase)
		expectedStatus int
		expectedNext   bool
	}{
		{
			name:        "successful cursor pagination - first page",
			queryParams: "?limit=2",
			setupMock: func(mockUC *MockWalletUseCase) {
				wallet := &models.Wallet{ID: 1, UserID: 1}
				mockUC.On("GetWalletByUserID", uint(1)).Return(wallet, nil)

				transactions := []models.Transaction{
					{ID: 3, CreatedAt: time.Now(), TransactionType: models.TransactionTypeCredit},
					{ID: 2, CreatedAt: time.Now().Add(-time.Hour), TransactionType: models.TransactionTypeDebit},
				}

				nextCursor := createTestCursor(2, time.Now().Add(-time.Hour))
				mockUC.On("GetTransactionHistory", uint(1), (*string)(nil), 2).
					Return(transactions, &nextCursor, nil)
			},
			expectedStatus: http.StatusOK,
			expectedNext:   true,
		},
		{
			name:        "successful cursor pagination - with cursor",
			queryParams: fmt.Sprintf("?cursor=%s&limit=2", createTestCursor(2, time.Now().Add(-time.Hour))),
			setupMock: func(mockUC *MockWalletUseCase) {
				wallet := &models.Wallet{ID: 1, UserID: 1}
				mockUC.On("GetWalletByUserID", uint(1)).Return(wallet, nil)

				transactions := []models.Transaction{
					{ID: 1, CreatedAt: time.Now().Add(-2 * time.Hour), TransactionType: models.TransactionTypeCredit},
				}

				// No next cursor means last page
				mockUC.On("GetTransactionHistory", uint(1), mock.MatchedBy(func(cursor *string) bool {
					return cursor != nil && *cursor != ""
				}), 2).
					Return(transactions, (*string)(nil), nil)
			},
			expectedStatus: http.StatusOK,
			expectedNext:   false,
		},
		{
			name:        "invalid direction parameter",
			queryParams: "?direction=invalid",
			setupMock: func(mockUC *MockWalletUseCase) {
				wallet := &models.Wallet{ID: 1, UserID: 1}
				mockUC.On("GetWalletByUserID", uint(1)).Return(wallet, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockWalletUseCase)
			tt.setupMock(mockUC)

			handler := NewWalletHandler(mockUC)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(1)) // Mock authenticated user
				c.Next()
			})
			router.GET("/wallets/me/transactions", handler.GetTransactionHistory)

			req, _ := http.NewRequest("GET", "/wallets/me/transactions"+tt.queryParams, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var response dto.APIResponse
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)

				// Check pagination metadata
				responseData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)

				pagination, ok := responseData["pagination"].(map[string]interface{})
				assert.True(t, ok)

				assert.Equal(t, tt.expectedNext, pagination["has_next_page"])
			}

			mockUC.AssertExpectations(t)
		})
	}
}
