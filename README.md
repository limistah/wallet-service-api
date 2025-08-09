# Wallet Service

A robust wallet service built with Go, providing comprehensive wallet management, transaction processing, and financial reconciliation capabilities.

## 🚀 Features

- **Wallet Management**: Create, manage, and monitor digital wallets
- **Transaction Processing**: Fund, withdraw, and transfer operations
- **Financial Reconciliation**: Automated balance verification and mismatch detection
- **Security**: Built-in validation and fraud prevention
- **Scalability**: Clean architecture with repository pattern

## 📋 Prerequisites

Before running this project, ensure you have the following installed:

- **Go**: Version 1.19 or higher
- **PostgreSQL**: Version 12 or higher (for production)
- **Git**: For version control

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd voyatek-test
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Verify installation**
   ```bash
   go version
   ```

## ⚙️ Environment Configuration

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=wallet_user
DB_PASSWORD=your_password
DB_NAME=wallet_service
DB_SSLMODE=disable

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY=24h

# Application Configuration
APP_ENV=development
LOG_LEVEL=info

# System Account Configuration
SYSTEM_ACCOUNT_EMAIL=system@walletservice.com
```

### Database Setup

1. **Create PostgreSQL database**
   ```sql
   CREATE DATABASE wallet_service;
   CREATE USER wallet_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE wallet_service TO wallet_user;
   ```

2. **Run database migrations** (if available)
   ```bash
   # Database migrations will be automatically handled by the application
   ```

## 🏃‍♂️ Running the Project

1. **Set up environment variables**
   ```bash
   export $(cat .env | xargs)
   ```

2. **Run the application**
   ```bash
   go run cmd/main.go
   ```

3. **Alternative: Use Makefile (if available)**
   ```bash
   make run
   ```

## 📊 API Documentation

Once the server is running, you can access:

- **API Documentation**: `http://localhost:8080/swagger/`
- **Health Check**: `http://localhost:8080/health`

## 🧪 Testing

This project includes comprehensive unit tests for all major components.

### Run All Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -v ./... -cover
```

## 🔧 Development

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

- Create an issue in the GitHub repository
- Check the troubleshooting section above
- Review the API documentation at `/swagger/`

---

**Current Test Coverage**: 32.9% (focused on business logic validation)

**Total Test Cases**: 31 individual test scenarios across wallet operations and reconciliation functionality
