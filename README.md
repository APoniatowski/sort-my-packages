# Pack Size Calculator API

This project implements a RESTful API for calculating optimal pack distributions for given quantities of items.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Git

### Installation

1. Clone the repository:

   ```
   git clone https://github.com/APoniatowski/sort-my-packages.git
   cd pack-size-calculator
   ```

2. Build the project:

   ```
   go build
   ```

3. Run the server:

   ```
   ./sort-my-packages
   ```

The server will start on `http://localhost:8080`.

## Usage

### API Endpoints

1. **Calculate Packs**
   - URL: `/calculate-packs`
   - Method: `POST`
   - Body: `{ "quantity": <number> }`
   - Response: `{ "packs": { "<pack_size>": <count>, ... }, "total_packs": <number> }`

2. **Set Pack Sizes**
   - URL: `/set-pack-sizes`
   - Method: `POST`
   - Headers:
     - `Content-Type: application/json`
     - `Authorization: Bearer <auth_token>`
     - `Origin: http://localhost:8080`
   - Body: `{ "pack_sizes": [<size1>, <size2>, ...] }`
   - Response: `"Pack sizes updated successfully"`

### Example Requests

Calculate packs for 1200 items:

```bash
curl -X POST http://localhost:8080/calculate-packs \
     -H "Content-Type: application/json" \
     -d '{"quantity": 1200}'
```

Set new pack sizes:

```bash
curl -X POST http://localhost:8080/set-pack-sizes \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer my_secret_token" \
     -H "Origin: http://localhost:8080" \
     -d '{"pack_sizes": [100, 200, 300, 400, 500]}'
```

## Configuration

The following configuration variables can be found in `calculatePacks.go`:

- `PackSizes`: Default pack sizes (can be updated via API)
- `AuthToken`: Authentication token for setting pack sizes
- `AllowedOrigin`: Allowed origin for CORS
- `MaxDPQuantity`: Maximum quantity for dynamic programming algorithm

## Testing

Run the test suite:

```
go test ./...
```

Execute the smoke test script:

```
./smoke-test.sh
```

## Makefile Usage

This project includes a Makefile to simplify common development tasks. Here are the available commands:

- `make build`: Builds the Go application.
- `make run`: Builds and runs the application locally.
- `make test`: Runs the test suite.
- `make docker-build`: Builds a Docker image for the application.
- `make docker-run`: Builds the Docker image and runs it in a container.
- `make docker-stop`: Stops and removes the running Docker container.
- `make clean`: Removes build artifacts.

### Examples

Build the application:

```
make build
```

Run the application locally:

```
make run
```

Build and run the Docker container:

```
make docker-run
```

Stop the Docker container:

```
make docker-stop
```

### Configuration

The Makefile uses the following variables, which you can modify if needed:

- `APP_NAME`: The name of the application (default: sort-my-packages)
- `BUILD_DIR`: The directory where the built binary will be placed (default: ./build)
- `MAIN_FILE`: The path to the main Go file (default: ./cmd/sort-my-packages/main.go)
- `DOCKER_IMAGE`: The name and tag for the Docker image (default: sort-my-packages:latest)
- `PORT`: The port on which the Docker container will run (default: 8080)

## Algorithms used

The project uses two algorithms for pack calculation:

1. **Dynamic Programming**: For quantities up to `MaxDPQuantity`, providing optimal solutions.
2. **Greedy Fallback**: For larger quantities, providing fast approximate solutions.

The algorithm considers factors such as minimizing overpack, total number of packs, and largest pack size based on the quantity.

