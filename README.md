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

### Development Mode

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

### Production Mode

1. **Build the application**
   ```bash
   go build -o bin/wallet-service cmd/main.go
   ```

2. **Run the binary**
   ```bash
   ./bin/wallet-service
   ```

### Using Docker (if Dockerfile is available)

1. **Build Docker image**
   ```bash
   docker build -t wallet-service .
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose up -d
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

### Adding New Features

1. **Create feature branch**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Write tests first** (TDD approach)
   ```bash
   # Add tests in appropriate test files
   go test -v ./internal/usecases -run="TestNewFeature"
   ```

3. **Implement feature**
   ```bash
   # Implement the feature
   go test ./... # Ensure all tests pass
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/new-feature
   ```

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check if PostgreSQL is running
   sudo systemctl status postgresql
   
   # Verify database exists
   psql -h localhost -U wallet_user -d wallet_service
   ```

2. **Port Already in Use**
   ```bash
   # Find process using port 8080
   lsof -i :8080
   
   # Kill process
   kill -9 <PID>
   ```

3. **Environment Variables Not Loaded**
   ```bash
   # Verify .env file exists and is readable
   cat .env
   
   # Source environment variables
   source .env
   ```

4. **Test Failures**
   ```bash
   # Run specific failing test
   go test -v ./internal/usecases -run="TestSpecificFunction"
   
   # Clear test cache
   go clean -testcache
   ```

### Debug Mode

```bash
# Run with debug logging
LOG_LEVEL=debug go run cmd/main.go

# Run tests with verbose output
go test -v ./... -count=1
```

## 📈 Performance

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkWalletOperation ./internal/usecases
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./...

# Memory profiling
go test -memprofile=mem.prof -bench=. ./...

# View profiles
go tool pprof cpu.prof
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Implement your changes
5. Ensure all tests pass
6. Submit a pull request

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
