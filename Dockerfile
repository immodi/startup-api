# Dockerfile
FROM golang:1.22.2-alpine AS build
WORKDIR /app

# Add necessary build deps
RUN apk add --no-cache gcc musl-dev

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download -x

# Copy source
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:3.19
WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/main .

# Make binary executable
RUN chmod +x /app/main

CMD ["./main", "serve", "--http=0.0.0.0:8090"]
