FROM golang:1.24-alpine as builder

# Install git and other necessary build tools
RUN apk update && apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy only the go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download && go mod tidy

# Copy the rest of the application
COPY . .

# Ensure all dependencies are properly resolved
RUN go mod tidy && go mod verify

WORKDIR /app/cmd/app

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /app/main .

# Use a smaller image for the final container
FROM alpine:latest

# Add CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy only the necessary files from the builder stage
# NOTE: .env is intentionally NOT copied — pass env vars via --env-file or docker compose
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

# Set the entry point to run the application
ENTRYPOINT ["./main"]