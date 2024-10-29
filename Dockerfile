# Dockerfile
FROM golang:1.22.2-alpine AS build
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy only the dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with memory and CPU constraints
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a minimal image for the final container
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
CMD ["./main", "serve", "--http=0.0.0.0:8090"]
