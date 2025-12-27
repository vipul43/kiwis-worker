# ---------- Build stage ----------
# Use official Go 1.25 image based on Alpine Linux
# This stage is ONLY for building the app
FROM golang:1.25-alpine AS builder

# Alpine images do not guarantee `make` is present
# Since our workflow depends on Makefile, we install make explicitly
RUN apk add --no-cache make git

# Set working directory inside the container to /app
# All following commands will run from this directory
WORKDIR /app

# Copy go.mod and go.sum first for better Docker layer caching
COPY go.mod go.sum ./

# Copy Makefile because we will use `make deps` and `make build`
COPY Makefile ./

# Download dependencies using Makefile abstraction
# Internally this runs `go mod download` and `go mod tidy`
RUN make deps

# Now copy the rest of the application source code
# This includes Go source files, configs, migrations, etc.
COPY . .

# Build the production-ready binary
# Internally this runs `go build -o bin/kiwis-worker cmd/kiwis-worker/main.go`
RUN make build

# ---------- Runtime stage ----------
# Minimal Alpine image for running the app
# No build tools or temp files from previous stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and make for potential runtime commands
RUN apk add --no-cache ca-certificates make

# Set working directory
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/bin/kiwis-worker ./kiwis-worker

# Copy migrations directory and Makefile for potential runtime operations
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/Makefile ./

# Document that the app might listen on port 8080
# (adjust based on your actual app configuration)
EXPOSE 8080

# Default command when the container starts
# Run the binary directly since it's already built
CMD ["./kiwis-worker"]