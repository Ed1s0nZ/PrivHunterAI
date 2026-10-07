.PHONY: build test dev clean
build:
	cd frontend && npm ci && npm run build
	go build -o bin/privhunter ./cmd/privhunter
test:
	go test -race ./...
	go vet ./...
	cd frontend && npm run build
dev:
	go run ./cmd/privhunter
clean:
	rm -rf bin frontend/dist
