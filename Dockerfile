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

CMD ["./main", "serve", "--http=0.0.0.0:8090"]