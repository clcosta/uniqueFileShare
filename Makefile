.PHONY: all
all: help

.PHONY: build
build:
	@go build -o ./build/shareFile ./cmd/shareFile/

.PHONY: run
run: build
	@export CGO_ENABLED=1
	@./build/shareFile

.PHONY: dev
dev:
	@go run ./cmd/shareFile/

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build   - Build the project"
	@echo "  run     - Build and run the project"
	@echo "  dev     - Run the project in development mode"
	@echo "  help    - Display this help message"
	@echo ""
	@echo "For more information, see the README.md file."
	@echo ""