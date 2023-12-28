# Use the official Golang image as the base image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application files to the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port your application will run on
EXPOSE 8086

# Command to run the application
CMD ["./main"]