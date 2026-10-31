# Build the API and operational commands from the same source revision.
FROM golang:1.27.1-trixie AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/seed-admin ./cmd/seed-admin

FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app -h /home/app app
WORKDIR /app
COPY --from=build --chown=app:app /out/api /usr/local/bin/api
COPY --from=build --chown=app:app /out/migrate /usr/local/bin/migrate
COPY --from=build --chown=app:app /out/seed-admin /usr/local/bin/seed-admin
COPY --from=build --chown=app:app /src/migrations ./migrations
USER app
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["api"]
