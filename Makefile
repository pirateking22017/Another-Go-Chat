.PHONY: build run clean test

# Variables
BINARY_NAME=chat-server
SOURCE_FILES=main.go private_message.go

# Build the application
build:
	@echo "Building chat server..."
	go build -o $(BINARY_NAME) $(SOURCE_FILES)

# Run the application
run:
	@echo "Running chat server..."
	go run main.go private_message.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Help command
help:
	@echo "Available commands:"
	@echo "  make build    - Build the chat server"
	@echo "  make run      - Run the chat server"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Install dependencies"
	@echo "  make help     - Show this help message" 