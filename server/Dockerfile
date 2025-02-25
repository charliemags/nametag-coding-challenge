FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy all necessary source files into the builder
# (Assuming you have server.go, main.go, latest.json, etc. in the same dir)
COPY . .

# Build the "server" binary that will serve files on port 8081
RUN go build -o server server.go

# Now cross-compile your 3 client binaries
# (Adjust architecture as needed; example is x86_64 for all)
RUN CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o myapp-darwin main.go
RUN CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o myapp-linux main.go
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o myapp-windows.exe main.go

FROM alpine:latest

WORKDIR /app

# Copy the server binary
COPY --from=builder /app/server /app/

# Copy the cross-compiled client binaries
COPY --from=builder /app/myapp-darwin    /app/
COPY --from=builder /app/myapp-linux     /app/
COPY --from=builder /app/myapp-windows.exe /app/

# Copy latest.json (so the server can serve it)
COPY --from=builder /app/latest.json     /app/

# We will listen on port 8201 internally, which can be published/redirected
EXPOSE 8201

# Start the server on port 8201
CMD ["./server", "-port=8201"]