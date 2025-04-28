# Download dependencies.
deps:
    go mod download

# Run all tests with race detection, verbose output, and atomic coverage tracking.
test:
    go test -race -v -covermode=atomic ./...
