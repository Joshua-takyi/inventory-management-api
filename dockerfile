# Stage 1: Build the Go application using the required Go version.
FROM golang:1.27.2-alpine AS build

# Set the working directory inside the build container.
WORKDIR /app

# Copy dependency manifests first to cache downloaded modules.
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compile a static Linux binary.
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# Stage 2: Use a minimal, non-root runtime image for security and size.
FROM gcr.io/distroless/static-debian12:nonroot

# Copy only the compiled binary; Go tooling is not needed at runtime.
COPY --from=build /api /api

# Run Gin in production mode to disable debug output.
ENV GIN_MODE=release

# Document the port the API is expected to listen on.
EXPOSE 8080

# Start the API when the container launches.
ENTRYPOINT ["/api"]