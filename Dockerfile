# server/Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy server source
COPY /server/server.go .
COPY /server/latest.json .

# Build the 'server' binary
RUN go build -o server server.go

# ---------------------------------------------------------
# Final Image
# ---------------------------------------------------------
FROM alpine:latest

WORKDIR /app

# Copy 'server' binary & latest.json
COPY --from=builder /app/server /app/
COPY --from=builder /app/latest.json /app/

# Copy the PRE-BUILT client binaries from your local repo
# (They are *not* built by Docker, just copied in.)
COPY /dist/myapp-darwin      /app/
COPY /dist/myapp-linux       /app/
COPY /dist/myapp-windows.exe /app/

EXPOSE 8201

CMD ["./server", "-port=8201"]
