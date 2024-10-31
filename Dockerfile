# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app

# Add necessary build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download -x

# Copy source code
COPY . .

# Build with stack size optimization
RUN CGO_ENABLED=0 GOOS=linux \
    GOARCH=amd64 \
    go build \
    -o main \
    -ldflags="-s -w" \
    -gcflags=all="-l" \
    .

# Runtime stage
FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install Chromium and dependencies
RUN apt-get update && apt-get install -y \
    chromium-browser \
    chromium-chromedriver \
    xvfb \
    ca-certificates \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libasound2 \
    libpangocairo-1.0-0 \
    libxdamage1 \
    libxshmfence1 \
    && rm -rf /var/lib/apt/lists/*

# Set up Xvfb for headless browser support
ENV DISPLAY=:99

# Set Go runtime environment variables
ENV GOGC=off
ENV GOMEMLIMIT=512MiB

# Create app directory
WORKDIR /app

# Copy the built executable from build stage
COPY --from=build /app/main .

# Copy the templates directory, .env file, and template.html into the /app directory in the runtime container
COPY templates/ /app/templates/
COPY .env /app/.env
COPY template.html /app/template.html

CMD ["./main", "serve", "--http=0.0.0.0:8090"]

