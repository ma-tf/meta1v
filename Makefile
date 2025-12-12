.PHONY: all
all: tidy generate lint build test

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test -race -v ./...

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

.PHONY: build
build:
	go build -o ./bin/meta1v .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run ./...

.PHONY: launch
launch:
	go run main.go


