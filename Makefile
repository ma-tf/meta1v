.PHONY: all
all: tidy generate lint build test docs

.PHONY: generate
generate:
	go generate ./...

PKGS := $(shell go list ./... 2>/dev/null | grep -Ev '(/test|/mocks|/cmd$$|/tools/|/osfs$$|/osexec$$|^github\.com/ma-tf/meta1v$$)')

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

.PHONY: update
update:
	go get -u ./...
	go mod tidy

.PHONY: docs
docs:
	@echo "Generating CLI documentation..."
	@go run internal/tools/docgen/main.go --out ./docs --format markdown
	@echo "CLI documentation generated in ./docs"

.PHONY: man
man:
	@echo "Generating man pages..."
	@go run internal/tools/docgen/main.go --out ./man --format man
	@echo "Man pages generated in ./man"

.PHONY: notice
notice:
	@echo "Generating NOTICE file..."
	@go-licenses report . --template=notice.tpl --include_tests > NOTICE
	@echo "NOTICE file updated"
