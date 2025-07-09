# MMad Stablecoin Main Makefile

# Variables
-include .env
export

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BUILD_DIR=build
TOOLS_DIR=tools

# Default target
.PHONY: help
help:
	@echo "MMad Stablecoin Build System"
	@echo "Available commands:"
	@echo "  make all             - Build everything (contracts + tools)"
	@echo "  make contracts       - Build Solidity contracts"
	@echo "  make tools           - Build Go tools"
	@echo "  make circuits        - Compile ZK circuits"
	@echo "  make test            - Run all tests"
	@echo "  make deploy-testnet  - Deploy to BSC testnet"
	@echo "  make clean           - Clean all artifacts"

# Build everything
.PHONY: all
all: contracts tools circuits

# Solidity contracts
.PHONY: contracts
contracts:
	@echo "Building Solidity contracts..."
	forge build

# Go tools
.PHONY: tools
tools:
	@echo "Building Go tools..."
	cd $(TOOLS_DIR) && make build

# ZK circuits
.PHONY: circuits
circuits:
	@echo "Compiling ZK circuits..."
	cd circom-tools && npm run compile

# Testing
.PHONY: test
test: test-contracts test-tools

.PHONY: test-contracts
test-contracts:
	@echo "Testing Solidity contracts..."
	forge test -vvv

.PHONY: test-tools
test-tools:
	@echo "Testing Go tools..."
	cd $(TOOLS_DIR) && make test

# Deployment
.PHONY: deploy-testnet
deploy-testnet:
	forge script script/Deploy.s.sol --rpc-url $(BSC_TESTNET_URL) --broadcast --verify

# Cleaning
.PHONY: clean
clean:
	forge clean
	cd $(TOOLS_DIR) && make clean
	rm -rf $(BUILD_DIR)

# Setup
.PHONY: install
install:
	@echo "Installing dependencies..."
	forge install
	cd circom-tools && npm install
	cd $(TOOLS_DIR) && go mod tidy

# Start services
.PHONY: start-monitor
start-monitor: tools
	./$(BUILD_DIR)/reserve-monitor

.PHONY: start-proof-service
start-proof-service: tools
	./$(BUILD_DIR)/proof-generator

# Docker
.PHONY: docker-build
docker-build:
	cd $(TOOLS_DIR) && make docker-build

.PHONY: docker-run
docker-run:
	cd $(TOOLS_DIR) && make docker-run