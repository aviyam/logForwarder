# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o logforwarder main.go

# Run stage
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/logforwarder .

# Expose ports
EXPOSE 5044 24224

# Set environment variables (defaults)
ENV LUMBERJACK_ADDR=":5044"
ENV FLUENT_ADDR=":24224"

# Command to run
CMD ["./logforwarder"]
