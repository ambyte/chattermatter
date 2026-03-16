# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /mattermost

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    curl \
    ca-certificates \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy server directory
COPY server/ ./

# Create go.work file to link modules (required for cyclic dependency resolution)
RUN go work init && \
    go work use . && \
    go work use ./public

# Download dependencies for both modules
RUN go mod download && \
    cd public && go mod download

# Set build environment
ENV CGO_ENABLED=1
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Install cross-compiler for ARM64 if needed
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        apt-get update && apt-get install -y gcc-aarch64-linux-gnu && \
        export CC=aarch64-linux-gnu-gcc; \
    fi

# Build binaries using go build directly
# We build only cmd/mattermost and cmd/mmctl
RUN CC=$(if [ "${TARGETARCH}" = "arm64" ]; then echo "aarch64-linux-gnu-gcc"; else echo "gcc"; fi) \
    go build -v \
    -o bin/mattermost \
    -trimpath \
    -tags 'production' \
    -ldflags '-w -s -X github.com/mattermost/mattermost/server/public/model.BuildNumber=docker' \
    ./cmd/mattermost && \
    CC=$(if [ "${TARGETARCH}" = "arm64" ]; then echo "aarch64-linux-gnu-gcc"; else echo "gcc"; fi) \
    go build -v \
    -o bin/mmctl \
    -trimpath \
    -tags 'production' \
    -ldflags '-w -s' \
    ./cmd/mmctl

# Verify binaries were created
RUN ls -la bin/

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    curl \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1000 mattermost && \
    useradd -u 1000 -g mattermost -d /mattermost mattermost

WORKDIR /mattermost

# Copy binaries from builder
COPY --from=builder /mattermost/bin/mattermost /mattermost/bin/mattermost
COPY --from=builder /mattermost/bin/mmctl /mattermost/bin/mmctl

# Copy necessary files from builder
COPY --from=builder /mattermost/i18n /mattermost/i18n
COPY --from=builder /mattermost/templates /mattermost/templates

# Create directories with proper permissions
# Mattermost needs write access to config, data, logs, plugins
RUN mkdir -p /mattermost/config /mattermost/data /mattermost/logs /mattermost/plugins /mattermost/client/plugins && \
    chown -R mattermost:mattermost /mattermost && \
    chmod -R u+w /mattermost/config /mattermost/data /mattermost/logs /mattermost/plugins /mattermost/client/plugins

# Switch to non-root user
USER mattermost

# Expose ports
EXPOSE 8065
EXPOSE 8067

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8065/api/v4/system/ping || exit 1

# Set environment variables
ENV PATH="/mattermost/bin:${PATH}"
ENV DISABLE_LICENSE="1"
ENV MM_CONFIG="/mattermost/config/config.json"
ENV MM_SERVICESETTINGS_ENABLELOCALMODE="true"

# Start mattermost
CMD ["mattermost"]
