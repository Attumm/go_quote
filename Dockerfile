FROM golang:1.23.3-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o go_quote

# Stage 2: Create the final minimal image
FROM scratch

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/go_quote .

# Copy the data file
COPY data/quotes.bytesz data/quotes.bytesz

# Set GOMAXPROCS environment variable
ENV GOMAXPROCS=1

# Set the entrypoint
ENTRYPOINT ["./go_quote", "-FILENAME", "data/quotes.bytesz", "-STORAGE", "bytesz"]

