# Stage 1: Build the application
FROM golang:1.22.0-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# -ldflags="-w -s" reduces the size of the binary by removing debug information.
# CGO_ENABLED=0 ensures a statically linked binary.
RUN CGO_ENABLED=0 go build -v -o /url-shortener -ldflags="-w -s" main.go

# Stage 2: Create the runtime image
FROM alpine:latest

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /url-shortener /app/url-shortener

# Copy .env.example, the actual .env file should be mounted or variables provided
# COPY .env.example .env.example # Optional: for reference inside the container

# Expose port 8080 (or the one configured via PORT env var)
EXPOSE 8080

# Set the user to the non-root user
USER appuser

# Command to run the executable
CMD ["./url-shortener"]
