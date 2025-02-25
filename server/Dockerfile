# server/Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy your Go server files into the container (this example assumes server.go, latest.json, etc. are in the same directory)
COPY . .

# Build the server binary
RUN go build -o server server.go

# Final stage (could also be multi-stage to reduce image size, but let's keep it simple)
FROM alpine:latest
WORKDIR /app

# Copy the server binary from the builder stage
COPY --from=builder /app/server /app/

# Copy other required files (like latest.json, or any binaries you host)
COPY --from=builder /app/latest.json /app/

EXPOSE 8080
CMD ["./server", "-port=8081"]