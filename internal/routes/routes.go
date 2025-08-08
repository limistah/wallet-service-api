package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/limistah/wallet-service/internal/auth"
	"github.com/limistah/wallet-service/internal/handlers"
	"github.com/limistah/wallet-service/internal/middleware"
	"github.com/limistah/wallet-service/internal/usecases"
)

func SetupRoutes(router *gin.Engine, useCases *usecases.UseCases, jwtService *auth.JWTService) {
	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	authHandler := handlers.NewAuthHandler(useCases.User, jwtService)
	authGroup := router.Group("/api/v1")
	{
		authGroup.POST("/auth/register", authHandler.Register)
		authGroup.POST("/auth/login", authHandler.Login)
		authGroup.POST("/auth/refresh", middleware.AuthMiddleware(jwtService), authHandler.RefreshToken)
		authGroup.POST("/auth/change-password", middleware.AuthMiddleware(jwtService), authHandler.ChangePassword)
	}

	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(jwtService))
	{
		walletHandler := handlers.NewWalletHandler(useCases.Wallet)
		wallets := v1.Group("/wallets")
		{
			wallets.GET("/me", walletHandler.GetWallet)                          // Get authenticated user's wallet
			wallets.GET("/me/balance", walletHandler.GetWalletBalance)           // Get authenticated user's wallet balance
			wallets.POST("/me/fund", walletHandler.FundWallet)                   // Fund authenticated user's wallet
			wallets.POST("/me/withdraw", walletHandler.WithdrawFunds)            // Withdraw from authenticated user's wallet
			wallets.POST("/me/transfer", walletHandler.TransferFunds)            // Transfer from authenticated user's wallet
			wallets.GET("/me/transactions", walletHandler.GetTransactionHistory) // Get authenticated user's transaction history
		}
	}
}
