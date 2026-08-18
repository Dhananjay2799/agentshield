FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/agentshield-alert-receiver \
    .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates \
       curl \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd \
      --gid 999 \
      agentshield \
    && useradd \
      --uid 999 \
      --gid 999 \
      --no-create-home \
      --shell /usr/sbin/nologin \
      agentshield

WORKDIR /app

COPY --from=builder \
  /out/agentshield-alert-receiver \
  /app/agentshield-alert-receiver

RUN chown agentshield:agentshield \
  /app/agentshield-alert-receiver

USER agentshield

EXPOSE 8084

ENTRYPOINT ["/app/agentshield-alert-receiver"]