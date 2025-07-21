# Build stage
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies and HTTPS)
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o vozdns .

# Final stage - scratch image for minimal size
FROM scratch

# Import ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Import user and group files from builder stage
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /build/vozdns /vozdns

# Use non-root user for security
USER appuser

# Expose port (if server mode is used)
EXPOSE 9000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/vozdns", "-help"] || exit 1

# Default command
ENTRYPOINT ["/vozdns"]
CMD ["-start"]

# Metadata
LABEL org.opencontainers.image.title="VozDNS"
LABEL org.opencontainers.image.description="Secure Dynamic DNS Client"
LABEL org.opencontainers.image.vendor="VozDNS"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/hypnguyen1209/vozdns"
