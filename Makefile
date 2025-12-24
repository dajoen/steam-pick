.PHONY: all build test lint clean run release-dry

BINARY_NAME=steam-pick
BUILD_DIR=bin

all: lint test build

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags="-X 'github.com/dajoen/steam-pick/internal/version.Version=$$(git describe --tags --always --dirty)' -X 'github.com/dajoen/steam-pick/internal/version.Commit=$$(git rev-parse HEAD)' -X 'github.com/dajoen/steam-pick/internal/version.Date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/steam-pick

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist

hooks:
	mkdir -p .git/hooks
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Hook installed."

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

release-dry:
	goreleaser release --snapshot --clean
