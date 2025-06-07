# Build stage
FROM golang:1.21-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates build-base

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o launchstack-api .

# Final stage
FROM alpine:3.18

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata docker-cli

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/launchstack-api .

# Copy environment files (will be overridden by docker-compose env_file)
COPY .env.example .env

# Create data directory for n8n
RUN mkdir -p /opt/n8n/data && chmod 777 /opt/n8n/data

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./launchstack-api"] 