# Build parameters
BINARY_NAME=wtfk8s
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
GO=go

# Go files
GO_FILES=$(shell find . -type f -name '*.go')

.PHONY: all build clean test lint install help

all: clean lint test build ## Run clean, lint, test and build

build: ## Build the binary
	${GO} build ${LDFLAGS} -o ${BINARY_NAME}

clean: ## Remove build artifacts
	${GO} clean
	rm -f ${BINARY_NAME}
	rm -f coverage.out

test: ## Run tests
	${GO} test -v -race -cover ./...

coverage: ## Run tests with coverage report
	${GO} test -v -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out

lint: ## Run linters
	golangci-lint run

install: ## Install binary to GOPATH
	${GO} install ${LDFLAGS}

fmt: ## Format code
	${GO} fmt ./...

vet: ## Run go vet
	${GO} vet ./...

mod-tidy: ## Tidy and verify go modules
	${GO} mod tidy
	${GO} mod verify

release: ## Create and push a new tag
	@echo "Creating release ${VERSION}"
	git tag -a ${VERSION} -m "Release ${VERSION}"
	git push origin ${VERSION}

docker: ## Build docker image
	docker build -t wtfk8s:${VERSION} .

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help 