FROM golang:1.20-alpine AS builder

# We'll create separate folders for server and client under /app
WORKDIR /app

# Copy the server folder (containing server.go, latest.json, etc.)
COPY server/ server/

# Copy the client folder (containing main.go, etc.)
COPY client/ client/

# Build output folder (we'll store final binaries here)
RUN mkdir -p /app/out

#
# Step 1: Build the "server" binary
#
WORKDIR /app/server
RUN go build -o /app/out/server server.go

#
# Step 2: Cross-compile the client binaries
# (Adjust architectures as needed)
#
WORKDIR /app/client

# macOS (Darwin) ARM64 (Apple Silicon) as example:
RUN CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o /app/out/myapp-darwin main.go

# Linux AMD64
RUN CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o /app/out/myapp-linux main.go

# Windows AMD64
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /app/out/myapp-windows.exe main.go

FROM alpine:latest

WORKDIR /app

# Copy the server binary
COPY --from=builder /app/out/server /app/

# Copy the cross-compiled client binaries
COPY --from=builder /app/out/myapp-darwin     /app/
COPY --from=builder /app/out/myapp-linux      /app/
COPY --from=builder /app/out/myapp-windows.exe /app/

# Copy latest.json (so the server can serve it)
COPY --from=builder /app/server/latest.json   /app/

# We'll expose port 8201 for the server
EXPOSE 8201

# Start the server on port 8201
CMD ["./server", "-port=8201"]
    