FROM golang:1.22.2 as build

WORKDIR /app

# Copy the Go module files
COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main", "serve", "--http=0.0.0.0:8090"]
