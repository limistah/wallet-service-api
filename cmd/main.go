package main

// @title Wallet Service API
// @version 1.0
// @description A wallet service API for managing user wallets and transactions
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"fmt"
	"log"
	"net/http"

	"github.com/limistah/wallet-service/docs"
	"github.com/limistah/wallet-service/internal/auth"
	"github.com/limistah/wallet-service/internal/config"
	"github.com/limistah/wallet-service/internal/database"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/limistah/wallet-service/internal/routes"
	"github.com/limistah/wallet-service/internal/usecases"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	cfg := config.LoadConfig()
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	repos := repositories.NewRepositories(db)

	useCases := usecases.NewUseCases(repos)

	jwtService := auth.NewJWTService(cfg.App.JWTSecret, "wallet-service")

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.SetupRoutes(router, useCases, jwtService)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("Server starting on %s:%s in %s mode",
		cfg.Server.Host, cfg.Server.Port, cfg.App.Environment)
	log.Printf("Swagger UI available at: http://%s:%s/swagger/index.html",
		cfg.Server.Host, cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
