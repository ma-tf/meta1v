.PHONY: build
build:
	go build ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: launch
launch:
	go run main.go
