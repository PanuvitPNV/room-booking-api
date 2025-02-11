# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Generate Swagger docs
RUN swag init -g cmd/main.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install curl for health check and postgresql-client for DB check
RUN apk add --no-cache curl postgresql-client

WORKDIR /app

# Copy wait-for script
COPY .docker/wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh

# Copy binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/internal/config/config.yaml /app/internal/config/

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 8080

# Use wait-for script
ENTRYPOINT ["/wait-for-postgres.sh"]
CMD ["./main"]