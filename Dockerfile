# Use an official Go runtime as a parent image
FROM golang:1.22-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the rest of the application code to the workspace
COPY . .

# Build the Go app
RUN go build -o /app/main .

# Expose port if your Go app is exposing any port (e.g., 8080)
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]
