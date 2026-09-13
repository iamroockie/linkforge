-include .env
export

MOCKGEN_VERSION := 0.6.0
GOLANGCI_VERSION := 2.13.1

BIN_DIR := bin
GOLANGCI := $(BIN_DIR)/golangci-lint
MOCKGEN := $(BIN_DIR)/mockgen

export PATH := $(PATH):$(CURDIR)/$(BIN_DIR)

.PHONY: prepare
prepare:
	@if [ ! -e .env ]; then cp .env.example .env; fi

.PHONY: run
run:
	@go run ./cmd/api

.PHONY: test
test:
	@go test ./...

.PHONY: coverage
coverage:
	@go test -coverprofile=coverage.out \
	    -coverpkg=$$(go list ./... | grep -v /.*test$ | paste -sd,) ./...
	@go tool cover -func=coverage.out | awk '/^total:/ {print $3}'
	@go tool cover -html=coverage.out
	@rm coverage.out

.PHONY: lint
lint: $(GOLANGCI)
	@$(GOLANGCI) run

.PHONY: format
format: $(GOLANGCI)
	@$(GOLANGCI) fmt

.PHONY: gen
gen: $(MOCKGEN)
	@go generate ./...

$(GOLANGCI):
	@mkdir -p $(BIN_DIR)
	@curl -sSfL https://golangci-lint.run/install.sh | \
        sh -s -- -b $(BIN_DIR) v$(GOLANGCI_VERSION)

$(MOCKGEN):
	@mkdir -p $(BIN_DIR)
	@GOBIN="$(CURDIR)/$(BIN_DIR)" go install go.uber.org/mock/mockgen@v$(MOCKGEN_VERSION)
