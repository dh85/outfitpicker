.PHONY: test test-short test-race test-coverage bench fuzz clean build install lint

# Test commands
test:
	go test -v ./...

test-short:
	go test -short -v ./...

test-race:
	go test -race -v ./...

test-coverage:
	./scripts/coverage.sh

# Benchmark and fuzz testing
bench:
	go test -bench=. -benchmem ./...

fuzz:
	go test -fuzz=FuzzCategoryFlow -fuzztime=30s ./internal/app

# Build commands
build:
	go build -o bin/outfitpicker ./cmd/outfitpicker
	go build -o bin/outfitpicker-admin ./cmd/outfitpicker-admin

install:
	go install ./cmd/outfitpicker
	go install ./cmd/outfitpicker-admin

# Quality assurance
lint:
	golangci-lint run

clean:
	rm -rf bin/ coverage.out coverage.html coverage.txt

# Development helpers
dev-setup:
	go mod tidy
	go mod download

# Release commands
release-snapshot:
	goreleaser release --snapshot --clean --skip=publish

release-check:
	goreleaser check

release-test:
	goreleaser release --snapshot --clean --skip=publish
	@echo "Testing generated binaries:"
	@chmod +x ./dist/outfitpicker_linux_amd64_v1/outfitpicker
	@./dist/outfitpicker_linux_amd64_v1/outfitpicker --version

all: clean lint test-coverage build