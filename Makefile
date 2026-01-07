.PHONY: all
all: tidy generate lint build test

.PHONY: generate
generate:
	go generate ./...

PKGS := $(shell go list ./... 2>/dev/null | grep -Ev '(/test|/mocks|/cmd$$|/osfs$$|^github\.com/ma-tf/meta1v$$)')

.PHONY: test
test:
	go test -race $(PKGS)

.PHONY: coverage
coverage:
	@go test -covermode=count -coverprofile=coverage.out $(PKGS)
	@go tool cover -html=coverage.out -o coverage.html

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


