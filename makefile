# Makefile for Voyatek Backend

# Variables
BINARY_NAME=main
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=cmd/main.go
AIR_CONFIG=.air.toml

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPKG=$(shell go list .)

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help run dev build clean test deps air-init install-air fmt vet lint

## help: Show this help message
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  ${GREEN}make dev${NC}         - Start development server with Air (auto-reload)"
	@echo "  ${GREEN}make run${NC}         - Run the application directly"
	@echo ""
	@echo "Build:"
	@echo "  ${GREEN}make build${NC}       - Build the application"
	@echo "  ${GREEN}make clean${NC}       - Clean build artifacts"
	@echo ""
	@echo "Dependencies:"
	@echo "  ${GREEN}make deps${NC}        - Download and install dependencies"
	@echo "  ${GREEN}make install-air${NC} - Install Air for development"
	@echo ""
	@echo "Code Quality:"
	@echo "  ${GREEN}make test${NC}        - Run tests"
	@echo "  ${GREEN}make fmt${NC}         - Format Go code"
	@echo "  ${GREEN}make vet${NC}         - Run go vet"
	@echo "  ${GREEN}make lint${NC}        - Run golint (requires golint to be installed)"
	@echo ""
	@echo "Air:"
	@echo "  ${GREEN}make air-init${NC}    - Initialize Air configuration"
	@echo ""

## dev: Start development server with Air (auto-reload)
dev: install-air clean
	@echo "${BLUE}Starting development server with Air...${NC}"
	@air

## run: Run the application directly
run: build
	@echo "${BLUE}Running application...${NC}"
	@./$(BINARY_PATH)

## build: Build the application
build: clean
	@echo "${BLUE}Building application...${NC}"
	@mkdir -p $(GOBIN)
	@go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "${GREEN}Build completed: $(BINARY_PATH)${NC}"

## clean: Clean build artifacts and temporary files
clean:
	@echo "${BLUE}Cleaning...${NC}"
	@rm -rf $(GOBIN)
	@rm -rf tmp/
	@rm -f build-errors.log
	@rm -f *.db
	@echo "${GREEN}Clean completed${NC}"

## test: Run tests
test:
	@echo "${BLUE}Running tests...${NC}"
	@go test -v -race -coverprofile=coverage.out ./...

## deps: Download and install dependencies
deps:
	@echo "${BLUE}Downloading dependencies...${NC}"
	@go mod download
	@go mod tidy
	@echo "${GREEN}Dependencies updated${NC}"

## install-air: Install Air for development
install-air:
	@echo "${BLUE}Checking Air installation...${NC}"
	@which air > /dev/null || (echo "${YELLOW}Installing Air...${NC}" && go install github.com/air-verse/air@latest)
	@echo "${GREEN}Air is ready${NC}"

## air-init: Initialize Air configuration
air-init:
	@echo "${BLUE}Initializing Air configuration...${NC}"
	@air init
	@echo "${GREEN}Air configuration created${NC}"

## fmt: Format Go code
fmt:
	@echo "${BLUE}Formatting code...${NC}"
	@go fmt ./...
	@echo "${GREEN}Code formatted${NC}"

## vet: Run go vet
vet:
	@echo "${BLUE}Running go vet...${NC}"
	@go vet ./...
	@echo "${GREEN}go vet completed${NC}"

## lint: Run golint (requires golint to be installed)
lint:
	@echo "${BLUE}Running golint...${NC}"
	@which golint > /dev/null || (echo "${RED}golint not installed. Install with: go install golang.org/x/lint/golint@latest${NC}" && exit 1)
	@golint ./...
	@echo "${GREEN}golint completed${NC}"

# Docker targets (optional)
## docker-build: Build Docker image
docker-build:
	@echo "${BLUE}Building Docker image...${NC}"
	@docker build -t wallet-service .
	@echo "${GREEN}Docker image built${NC}"

## docker-run: Run Docker container
docker-run:
	@echo "${BLUE}Running Docker container...${NC}"
	@docker run -p 3000:3000 wallet-service

# Database targets
## db-reset: Reset database (remove database file)
db-reset:
	@echo "${BLUE}Resetting database...${NC}"
	@rm -f app.db
	@echo "${GREEN}Database reset${NC}"

## docker-up: Start MySQL with Docker
docker-up:
	@echo "${BLUE}Starting MySQL with Docker...${NC}"
	@docker-compose up -d mysql
	@echo "${GREEN}MySQL started${NC}"

## docker-down: Stop Docker services
docker-down:
	@echo "${BLUE}Stopping Docker services...${NC}"
	@docker-compose down
	@echo "${GREEN}Docker services stopped${NC}"

## docker-test-up: Start MySQL for testing
docker-test-up:
	@echo "${BLUE}Starting MySQL for testing...${NC}"
	@docker-compose up -d mysql_test
	@echo "${GREEN}Test MySQL started${NC}"

## setup-env: Create .env file from example
setup-env:
	@echo "${BLUE}Setting up environment...${NC}"
	@cp .env.example .env
	@echo "${GREEN}.env file created from .env.example${NC}"
	@echo "${YELLOW}Please edit .env file with your configuration${NC}"

## docs: Generate Swagger documentation
docs:
	@echo "${BLUE}Generating Swagger documentation...${NC}"
	@swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "${GREEN}Swagger documentation generated${NC}"

## docs-serve: Generate and serve Swagger documentation
docs-serve: docs run

# Production targets
## prod-build: Build for production
prod-build:
	@echo "${BLUE}Building for production...${NC}"
	@mkdir -p $(GOBIN)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_PATH) $(CMD_PATH)
	@echo "${GREEN}Production build completed${NC}"

# Default target
.DEFAULT_GOAL := help
