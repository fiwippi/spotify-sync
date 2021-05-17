FROM golang:1.16.3-alpine3.13 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server cmd/server.go

FROM scratch

COPY --from=builder /app/bin/server /server
COPY --from=builder etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV DOCKER=true
EXPOSE 8096
CMD ["/server"]