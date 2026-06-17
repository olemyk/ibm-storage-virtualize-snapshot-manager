# Testing Guide

This document describes how to run and write tests for the IBM Storage Virtualize Snapshot Manager backend.

## Running Tests

### Using Make (Recommended)

The project includes a Makefile with convenient test commands:

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage report
make test-coverage

# Run only API tests
make test-api

# Show all available make commands
make help
```

### Using Go Commands Directly

```bash
# Run all tests
go test ./cmd/... ./internal/... ./pkg/...

# Run tests with verbose output
go test -v ./cmd/... ./internal/... ./pkg/...

# Run tests with coverage
go test -coverprofile=coverage.out ./cmd/... ./internal/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test -v ./internal/api/...
```

## Test Coverage

After running `make test-coverage`, open `coverage.html` in your browser to see detailed coverage information:

```bash
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

Current test coverage:
- `internal/api`: 1.2% (helpers.go fully tested)
- Other packages: 0% (no tests yet)

## Writing Tests

### Test File Naming

Test files should be named `*_test.go` and placed in the same directory as the code they test.

Example:
- Code: `internal/api/helpers.go`
- Tests: `internal/api/helpers_test.go`

### Test Function Naming

Test functions should start with `Test` followed by the function name being tested:

```go
func TestHandleError(t *testing.T) {
    // Test implementation
}
```

### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestRespondJSON(t *testing.T) {
    tests := []struct {
        name       string
        status     int
        data       interface{}
        wantStatus int
        wantBody   string
    }{
        {
            name:       "success with data",
            status:     http.StatusOK,
            data:       map[string]string{"message": "success"},
            wantStatus: http.StatusOK,
            wantBody:   `{"message":"success"}`,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Testing HTTP Handlers

Use `httptest` package for testing HTTP handlers:

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestMyHandler(t *testing.T) {
    // Create a request
    req := httptest.NewRequest("GET", "/test", nil)
    
    // Create a response recorder
    w := httptest.NewRecorder()
    
    // Call the handler
    myHandler(w, req)
    
    // Check the response
    if w.Code != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
    }
}
```

### Testing with Context

For testing functions that use context:

```go
func TestGetUserIDFromContext(t *testing.T) {
    r := httptest.NewRequest("GET", "/test", nil)
    ctx := context.WithValue(r.Context(), userIDKey, 123)
    r = r.WithContext(ctx)
    
    got := getUserIDFromContext(r)
    if got != 123 {
        t.Errorf("Expected 123, got %d", got)
    }
}
```

## Test Organization

### Current Test Files

- `internal/api/helpers_test.go` - Tests for API helper functions
  - `TestRespondJSON` - Tests JSON response helper
  - `TestRespondError` - Tests error response helper
  - `TestHandleError` - Tests centralized error handling
  - `TestGetUserIDFromContext` - Tests user ID extraction from context
  - `TestGetUsernameFromContext` - Tests username extraction from context

### Future Test Files (Recommended)

- `internal/api/handlers_test.go` - Tests for API handlers
- `internal/auth/auth_test.go` - Tests for authentication
- `internal/db/db_test.go` - Tests for database operations
- `internal/svc/client_test.go` - Tests for IBM SVC client
- `internal/scheduler/scheduler_test.go` - Tests for scheduler
- `pkg/crypto/crypto_test.go` - Tests for encryption/decryption

## Best Practices

1. **Test One Thing**: Each test should verify one specific behavior
2. **Use Descriptive Names**: Test names should clearly describe what they test
3. **Arrange-Act-Assert**: Structure tests with setup, execution, and verification
4. **Table-Driven Tests**: Use for testing multiple scenarios of the same function
5. **Mock External Dependencies**: Use interfaces and mocks for external services
6. **Test Error Cases**: Always test both success and failure scenarios
7. **Clean Up**: Use `defer` to clean up resources (close files, connections, etc.)

## Continuous Integration

To integrate tests into CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run tests
  run: |
    cd backend
    make test-coverage
```

## Troubleshooting

### Tests Fail to Run

If you see "main redeclared" errors, ensure you're not running tests on the `scripts/` directory:

```bash
# Correct - excludes scripts
go test ./cmd/... ./internal/... ./pkg/...

# Incorrect - includes scripts with multiple main functions
go test ./...
```

### Coverage Report Not Generated

Ensure you have write permissions in the backend directory and that `go tool cover` is available:

```bash
go tool cover -h
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go HTTP Test Package](https://pkg.go.dev/net/http/httptest)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)