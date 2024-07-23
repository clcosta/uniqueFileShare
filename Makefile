.PHONY: build
build:
	@go build -o ./build/shareFile ./cmd/shareFile/

.PHONY: run
run: build
	@export CGO_ENABLED=1
	@./build/shareFile