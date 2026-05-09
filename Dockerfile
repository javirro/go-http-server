# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download modules first (cached layer when source hasn't changed).
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/bin/server ./cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM scratch

# Import CA certificates for outbound HTTPS calls.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the binary.
COPY --from=builder /app/bin/server /server

# Non-root user (numeric UID required with scratch).
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
