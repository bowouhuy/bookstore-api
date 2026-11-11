GOPATH=$(shell go env GOPATH)

GOLANGCI_LINT_VERSION=v1.64.8
COMPOSE=docker compose -f development/docker-compose.yml

install: install-gomock install-linter

install-gomock:
	@echo "\n>>> Install gomock\n"
	go install go.uber.org/mock/mockgen@v0.6.0

install-linter:
	@echo "\n>>> Install GolangCI-Lint"
	@echo ">>> https://github.com/golangci/golangci-lint/releases \n"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh | \
	sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}

lint:
	@echo "\n>>> Run GolangCI-Lint\n"
	/bin/bash ./scripts/lint.sh

test:
	mkdir -p .coverage/html
	go test -v -race -cover -coverprofile=.coverage/internal.coverage ./internal/... && \
	cat .coverage/internal.coverage | grep -v "_mock.go\|_mockgen.go" > .coverage/internal.mockless.coverage && \
	go tool cover -html=.coverage/internal.mockless.coverage -o .coverage/html/internal.coverage.html;

mock:
	@echo "\n>>> Run Generate Mock\n"
	go generate ./...

http:
	@go run main.go http:start

build:
	CGO_ENABLED=0 go build -o bin/bookstore-api .

db-up:
	$(COMPOSE) up -d db

migrate-up:
	$(COMPOSE) --profile tools run --rm migrate up

migrate-down:
	$(COMPOSE) --profile tools run --rm migrate down 1
