.PHONY: dev test test-coverage

dev:
	go run ./cmd/server/

test:
	go test -v -race -count=1 ./...

test-coverage:
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
