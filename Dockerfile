# Use the official Golang image to build the app
FROM golang:1.18 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Start a new stage from scratch
FROM alpine:latest  

# Install SQLite3
RUN apk --no-cache add sqlite

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main /app/main

# Copy database initialization script
COPY --from=builder /app/computers.db /app/computers.db

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/app/main"]
