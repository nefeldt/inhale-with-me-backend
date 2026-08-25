.PHONY: run build test tidy fmt vet docker-build

# Run the server locally (reads .env in development).
run:
	go run ./cmd/server

# Build a static binary.
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

# Build the production container image.
docker-build:
	docker build -t inhale-with-me-backend:local .
