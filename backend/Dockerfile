# Multi-stage build for minimal image size
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO disabled for static linking (removed -s flag to keep symbols for debugging)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w" -o snapshot-manager ./cmd/server

# Final stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/snapshot-manager .

# Copy embedded SQL schema (needed by db package)
COPY --from=builder /app/internal/db/schema.sql ./internal/db/
COPY --from=builder /app/internal/db/schema_postgres.sql ./internal/db/


# Create directories for logs and data
RUN mkdir -p /app/logs /app/data

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./snapshot-manager"]
