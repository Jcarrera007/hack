.PHONY: build run dev clean templ

# Build the application
build: templ
	go build -o bin/goodoo .

# Run the application
run: build
	./bin/goodoo

# Development mode with auto-reload
dev: templ
	go run .

# Generate templ templates
templ:
	templ generate

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f templates/*.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...
	templ fmt .

# Test the application
test:
	go test ./...

# Install templ CLI if not present
install-templ:
	go install github.com/a-h/templ/cmd/templ@latest