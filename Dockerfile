# Dockerfile
FROM golang:1.22.2-alpine AS build
WORKDIR /app

# Add necessary build deps
RUN apk add --no-cache gcc musl-dev

# Copy dependency files
COPY go.mod go.sum ./

# Download deps with limited concurrency
RUN go mod download -x

# Copy source
COPY . .

# Build with optimizations and constraints
RUN CGO_ENABLED=0 GOOS=linux \
    GOGC=50 \
    go build \
    -ldflags="-s -w" \
    -o main .

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/main .
RUN chmod +x /app/main

# Add basic tools and security updates
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -H -h /app appuser && \
    chown appuser:appuser /app/main

USER appuser

CMD ["./main", "serve", "--http=0.0.0.0:8090"]
