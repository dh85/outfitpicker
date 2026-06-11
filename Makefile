.PHONY: test build clean run coverage coverage-check ci

COVERAGE_MIN ?= 92.5

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

coverage-check:
	go test -coverprofile=coverage.out ./...
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	if awk 'BEGIN { exit !('"$$total"' + 0 >= '"$(COVERAGE_MIN)"' + 0) }'; then \
		echo "coverage $$total% meets threshold $(COVERAGE_MIN)%"; \
	else \
		echo "coverage $$total% is below threshold $(COVERAGE_MIN)%"; \
		exit 1; \
	fi

ci: coverage-check

build:
	go build -o bin/outfitpicker ./cmd/outfitpicker

clean:
	rm -rf bin/ coverage.out coverage.html

run: build
	./bin/outfitpicker
