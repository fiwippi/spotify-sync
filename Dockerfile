FROM golang:1.16.3-alpine3.13 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/spotify_sync

FROM scratch

COPY --from=builder /app/bin/spotify_sync /spotify_sync
COPY --from=builder etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8096
CMD ["/spotify_sync", "server", "--port", "8096"]