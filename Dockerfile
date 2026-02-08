# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cadence ./cmd/cadence

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata git

# Copy binary from builder
COPY --from=builder /app/cadence .

# Create config directory
RUN mkdir -p /root/.cadence

# Expose port for webhook server
EXPOSE 8000

# Default command - start webhook server
CMD ["./cadence", "webhook", "--host", "0.0.0.0", "--port", "8000"]
