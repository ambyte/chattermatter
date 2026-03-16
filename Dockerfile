# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make curl

WORKDIR /build

# Copy go mod files
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy source code
COPY server/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mattermost-server ./channels/cmd/mattermost

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S mattermost && \
    adduser -u 1000 -S mattermost -G mattermost

WORKDIR /mattermost

# Copy binary from builder
COPY --from=builder /build/mattermost-server /mattermost/bin/mattermost

# Copy necessary files
COPY --from=builder /build/channels/config/config.json /mattermost/config/config.json
COPY --from=builder /build/i18n /mattermost/i18n
COPY --from=builder /build/channels/templates /mattermost/templates

# Create data directories
RUN mkdir -p /mattermost/data /mattermost/logs /mattermost/plugins /mattermost/client/plugins && \
    chown -R mattermost:mattermost /mattermost

# Switch to non-root user
USER mattermost

# Expose ports
EXPOSE 8065
EXPOSE 8067
EXPOSE 8074

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8065/api/v4/system/ping || exit 1

# Set environment variables
ENV PATH="/mattermost/bin:${PATH}"
ENV MM_DISABLE_LICENSE="true"

# Start mattermost
CMD ["mattermost"]
