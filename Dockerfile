# Build stage
FROM golang:latest AS build

WORKDIR /app

# Copy go.mod and sum files first
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s -buildid=" -trimpath -o main .

# Runtime stage  
FROM debian:12.8-slim
WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/main .

CMD ["./main"]
