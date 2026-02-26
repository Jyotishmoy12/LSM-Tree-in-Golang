# Step 1: Build the Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .

ENV GOTOOLCHAIN=auto

RUN go mod download
# Build the TCP server we created earlier
RUN go build -o lsm-server ./cmd/lsm-server/main.go

# Step 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/
# Copy the binary from the builder
COPY --from=builder /app/lsm-server .
# Create a directory for persistent storage
RUN mkdir -p /root/stress_storage
# Expose the TCP port
EXPOSE 6379
# Run the server
CMD ["./lsm-server"]