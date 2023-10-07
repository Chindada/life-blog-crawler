BIN_NAME = lbcrawler

run: ### run
	@go mod tidy
	@go mod download
	@echo "Building $(BIN_NAME)..."
	@go build -o $(BIN_NAME) ./cmd/app
	@echo "Running $(BIN_NAME)..."
	@./$(BIN_NAME)
.PHONY: run

build: ### build
	@go mod tidy
	@go mod download
	@go build -o $(BIN_NAME) ./cmd/app
.PHONY: build

go-mod-update: ### go-mod-update
	@./scripts/gomod_update.sh
.PHONY: go-mod-update

update: go-mod-update ### update
.PHONY: update

lint: ### check by golangci linter
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run
.PHONY: lint

test: ### run test
	@go test ./... -v -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -func coverage.txt
.PHONY: test

help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help
