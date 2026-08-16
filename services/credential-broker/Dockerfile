FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/agentshield-credential-broker \
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
    /out/agentshield-credential-broker \
    /app/agentshield-credential-broker

RUN chown agentshield:agentshield \
    /app/agentshield-credential-broker

USER agentshield

EXPOSE 8081

ENTRYPOINT ["/app/agentshield-credential-broker"]