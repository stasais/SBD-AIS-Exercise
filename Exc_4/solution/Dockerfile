## Multi-stage Dockerfile copied from Exercise 3 and adapted for Exercise 4

# First stage (builder) - compile the application
FROM golang:1.25 AS builder

# Set working directory to /app
WORKDIR /app

# Copy all contents to /app directory
COPY . .

# Execute the build script (creates binary at /app/ordersystem)
RUN chmod +x scripts/build-application.sh \
    && ./scripts/build-application.sh

# Second stage - copy the compiled binary from builder and start the application
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the compiled binary from builder stage
COPY --from=builder /app/ordersystem ./ordersystem

# Expose application port
EXPOSE 3000

# Start the application on container run
CMD ["./ordersystem"]
