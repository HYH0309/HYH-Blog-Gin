# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.23-alpine AS build
WORKDIR /app

# Enable Go modules and install build deps
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Cache go.mod first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary (static)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20
WORKDIR /srv

# Add non-root user
RUN adduser -D -H -u 10001 appuser

# Certificates for HTTPS outbound
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Copy binary
COPY --from=build /bin/server /usr/local/bin/server

# Env
ENV PORT=8080
EXPOSE 8080

USER appuser
ENTRYPOINT ["/usr/local/bin/server"]
