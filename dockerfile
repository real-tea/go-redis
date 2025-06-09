# Stage 1: Build the Go binary
# We use a Go version that matches the go.mod file.
FROM golang:1.24-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies.
# This is done in a separate step to leverage Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Explicitly copy the Go source files into the container.
COPY src/hello/*.go ./

# --- DEBUGGING STEP ---
# List the contents of the /app directory to ensure files were copied.
RUN ls -la

# Build the Go application.
# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# Stage 2: Create the final, small production image
# We use a minimal base image. 'alpine' is a popular choice.
FROM alpine:latest

# It's good practice to run containers as a non-root user.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy only the compiled binary from the 'builder' stage.
COPY --from=builder /app/main /app/main

# Set the user for the container
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# The command to run when the container starts
CMD ["/app/main"]
