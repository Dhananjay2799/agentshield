FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/agentshield-audit \
    ./cmd/server


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
    /out/agentshield-audit \
    /app/agentshield-audit

RUN chown agentshield:agentshield \
    /app/agentshield-audit

USER agentshield

EXPOSE 8082

ENTRYPOINT ["/app/agentshield-audit"]