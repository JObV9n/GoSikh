# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

ENV CGO_ENABLED=1

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server

# Final stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app

RUN mkdir -p /app/data /app/sandbox/runners

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy any necessary configuration or static files
# COPY config.yaml ./

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]