FROM golang:1.25-alpine AS build

WORKDIR /app

# Cache module downloads as a separate layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api


# ---- Runtime stage ----
# Distroless: no shell, no package manager, runs as non-root by default.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /api /api

# Run Gin in release mode: no route-table dumps or debug warnings in logs.
ENV GIN_MODE=release

# Configuration is injected by the platform as env vars
# (PORT, DATABASE_URL, SUPABASE_*, PAYSTACK_SECRET_KEY, CORS_ORIGIN).
EXPOSE 8080
ENTRYPOINT ["/api"]