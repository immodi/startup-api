FROM golang:1.22.2 as build

# Create a non-root user and group
RUN groupadd -r appuser && useradd -r -g appuser appuser

WORKDIR /app

# Copy the Go module files
COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

COPY . .

# Build the application
RUN go build -o main .

# Set proper ownership and permissions
RUN chown appuser:appuser /app \
    && chmod 755 /app \
    && chown appuser:appuser /app/main \
    && chmod 755 /app/main

EXPOSE 8090

# Switch to non-root user
USER appuser

CMD ["./main", "serve", "--http=0.0.0.0:8090"]