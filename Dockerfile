# Use the official Golang image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git ca-certificates

# Copy all the files from your project folder into the container's /app folder
COPY . .

# Download Go modules
RUN go mod tidy

# Build the Go application.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main /app/cmd/server

# Make the binary executable
RUN chmod +x /app/main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]