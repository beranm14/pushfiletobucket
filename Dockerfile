FROM golang:latest AS build

WORKDIR /opt
COPY main.go /opt/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=0 /opt/main /main

ENV PORT 3000

EXPOSE 3000/tcp

ENTRYPOINT ["/main"]
