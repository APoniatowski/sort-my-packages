APP_NAME = sort-my-packages
BUILD_DIR = ./build
MAIN_FILE = ./cmd/sort-my-packages/main.go
DOCKER_IMAGE = $(APP_NAME):latest
PORT = 8080

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building the Go application..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete! Binary is located at $(BUILD_DIR)/$(APP_NAME)."

.PHONY: run
run: build
	@echo "Running it locally..."
	@$(BUILD_DIR)/$(APP_NAME)

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: docker-build
docker-build:
	@echo "Building the Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image $(DOCKER_IMAGE) built successfully."

.PHONY: docker-run
docker-run: docker-build
	@echo "Running the Docker container..."
	@docker run -d -p $(PORT):8080 --name $(APP_NAME) $(DOCKER_IMAGE)
	@echo "Docker container $(APP_NAME) is running on port $(PORT)."

.PHONY: docker-stop
docker-stop:
	@echo "Stopping and removing Docker container..."
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true

.PHONY: clean
clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup completed."
