# Start from golang:1.17-alpine base image
FROM golang:1.17-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies.
# Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app with name 'user' as the executable file
RUN go build -o user ./cmd

# Expose port 4000 to the outside world
EXPOSE 4000

# Run the executable file
CMD ["./user"]