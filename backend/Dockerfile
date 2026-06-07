# ---------- Build stage ----------
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Copy module file first for layer caching. go.sum may be absent on first build —
# `go mod tidy` below will create/sync it.
COPY go.mod ./
COPY go.sum* ./

# Copy the rest of the source.
COPY . .

# Resolve all imports and (re)generate go.sum. This fetches dependencies from
# proxy.golang.org and works whether go.sum existed or not.
RUN go mod tidy

# Build the two binaries.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---------- Runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S somon && adduser -S somon -G somon

WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /src/migrations /app/migrations

RUN mkdir -p /app/uploads && chown -R somon:somon /app

USER somon

EXPOSE 8080
ENTRYPOINT ["/app/server"]
