MODULE      := github.com/octamagic/user-settings
GENERATED   := internal/generated
UNIT_PKGS   := $(shell go list ./... | grep -v -e /generated -e /tests/)

.PHONY: help tools generate generate-proto generate-mocks build run test test-coverage test-integration test-e2e test-all lint docker-up docker-down tf-fmt tf-lint tf-test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

# The generators are built with the module's Go version, not whatever is on PATH:
# protoc-gen-twirp gzips the file descriptor at generation time and compress/flate
# output changes between Go releases, so a newer local Go produces a different
# (but equivalent) byte slice and the CI freshness check fails.
GO_VERSION := $(shell awk '/^go /{print $$2}' go.mod)

tools: ## Install the code generators, pinned, built with the go.mod Go version
	GOTOOLCHAIN=go$(GO_VERSION) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	GOTOOLCHAIN=go$(GO_VERSION) go install github.com/twitchtv/twirp/protoc-gen-twirp@v8.1.3
	GOTOOLCHAIN=go$(GO_VERSION) go install go.uber.org/mock/mockgen@v0.6.0

dev-tools: tools ## Generators plus the live-reload runner used by `make run`
	go install github.com/air-verse/air@latest

generate: generate-proto generate-mocks ## Regenerate protobuf, Twirp and mock code

generate-proto: ## Regenerate protobuf and Twirp code from rpc/*.proto
	rm -rf $(GENERATED) && mkdir -p $(GENERATED)
	protoc -I rpc --go_out=$(GENERATED) --go_opt=module=$(MODULE)/$(GENERATED) rpc/session.proto rpc/user_settings_api.proto
	protoc -I rpc --twirp_out=$(GENERATED) --twirp_opt=module=$(MODULE)/$(GENERATED) rpc/user_settings_api.proto

generate-mocks: ## Regenerate gomock mocks
	go generate ./...

build: ## Build the API binary into bin/
	CGO_ENABLED=0 go build -o bin/user-settings ./cmd/api

run: docker-up ## Run the API locally with live reload against DynamoDB Local
	set -a && . ./config/.env.local && set +a && air

test: ## Run unit tests
	go test -race -timeout 60s $(UNIT_PKGS)

test-coverage: ## Run unit tests with an HTML coverage report
	go test -race -timeout 60s -coverprofile coverage.out $(UNIT_PKGS)
	go tool cover -html coverage.out

test-integration: ## Run repository tests against DynamoDB Local (make docker-up first)
	set -a && . ./config/.env.test && set +a && go test -v -p 1 -timeout 120s ./tests/integration/...

test-e2e: ## Run end-to-end tests (starts containers via testcontainers)
	set -a && . ./config/.env.e2e && set +a && go test -v -p 1 -timeout 180s ./tests/e2e/...

test-all: docker-up test test-integration test-e2e ## Run every test suite

lint: ## Run golangci-lint
	golangci-lint run

docker-up: ## Start DynamoDB Local and the admin UI
	docker compose up -d --wait --wait-timeout 60

docker-down: ## Stop local containers
	docker compose down

tf-fmt: ## Format Terraform
	terraform fmt -recursive terraform

tf-lint: ## Lint Terraform with tflint
	tflint --init >/dev/null && tflint --recursive --format compact

tf-test: ## Validate and test Terraform against a local tf-tools checkout (TF_TOOLS_DIR)
	scripts/tf-local.sh
