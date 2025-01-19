# Use official golang image as builder
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum (if they exist)
COPY . .

# Build the application
RUN go build -o bot-trap .

# Use a minimal alpine image for the final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bot-trap .
# Copy static files
COPY index.html .
COPY robots.txt .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./bot-trap"]
