.PHONY: build run test docker-build docker-run docker-stop clean dev

# Build backend
build:
	CGO_ENABLED=0 go build -o uniops ./cmd/uniops

# Build frontend
build-frontend:
	cd web && npm run build

# Run locally
run: build
	./uniops

# Run tests
test:
	go test ./...

# Docker build
docker-build:
	docker-compose build

# Docker run
docker-run:
	docker-compose up -d

# Docker stop
docker-stop:
	docker-compose down

# Clean
clean:
	rm -f uniops uniops.exe
	rm -rf web/dist
	rm -rf data

# Development mode with hot reload
dev:
	go run ./cmd/uniops
