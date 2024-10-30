# Build stage
FROM golang:1.21-alpine AS builder

# Install required build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o pocketbase

# Final stage
FROM alpine:3.19

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a dedicated user and group with specific UID/GID
RUN addgroup -S pocketbase -g 1000 && \
    adduser -S pocketbase -u 1000 -G pocketbase

# Create necessary directories
WORKDIR /app
RUN mkdir -p /app/pb_data && \
    mkdir -p /app/pb_public && \
    mkdir -p /app/pb_migrations && \
    chown -R pocketbase:pocketbase /app && \
    chmod -R 744 /app

# Switch to non-root user
USER pocketbase:pocketbase

# Copy the binary from builder
COPY --from=builder --chown=pocketbase:pocketbase /app/pocketbase /app/pocketbase
RUN chmod +x /app/pocketbase

# Copy static files and migrations if they exist
COPY --chown=pocketbase:pocketbase pb_public/ /app/pb_public/
COPY --chown=pocketbase:pocketbase pb_migrations/ /app/pb_migrations/

# Expose PocketBase port
EXPOSE 8090

# Set environment variables
ENV PB_ENCRYPTION_KEY=""

# Command to run the application
CMD ["./pocketbase", "serve", "--http=0.0.0.0:8090"]