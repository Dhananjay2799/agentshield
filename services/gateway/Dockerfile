# ============================================================
# AgentShield Gateway
# Multi-stage production Docker build
# ============================================================

FROM golang:1.26-bookworm AS builder

WORKDIR /src

# Download dependencies separately so Docker can cache them.
COPY go.mod go.sum ./

RUN go mod download

# Copy Gateway source.
COPY . .

# Build a static Linux binary.
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/agentshield-gateway \
    ./cmd/server


# ============================================================
# Runtime
# ============================================================

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates \
       curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system agentshield \
    && useradd \
       --system \
       --gid agentshield \
       --no-create-home \
       --shell /usr/sbin/nologin \
       agentshield

WORKDIR /app

COPY --from=builder \
    /out/agentshield-gateway \
    /app/agentshield-gateway

RUN chown agentshield:agentshield \
    /app/agentshield-gateway

USER agentshield

EXPOSE 8080

ENTRYPOINT ["/app/agentshield-gateway"]