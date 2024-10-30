FROM golang:1.22-alpine AS build
WORKDIR /app

# Add necessary build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download -x

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final Stage
FROM alpine:3.19
WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/main .

# Set permissions for the /app directory and main executable
RUN chmod u+rwx /app/main && chmod -R u+rwx /app

CMD ["./main", "serve", "--http=0.0.0.0:8090"]