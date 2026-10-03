FROM golang:1.27.1-alpine3.24@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

WORKDIR $GOPATH/src/do-dyndns/app/
COPY app/go.mod .
COPY app/go.sum .

RUN go mod download

COPY app/ .

WORKDIR $GOPATH/src/do-dyndns/app/cmd/do-dyndns/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/do-dynds

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

RUN addgroup -S -g 10001 dyndns && adduser -S -u 10001 dyndns -G dyndns

COPY --from=builder /go/bin/do-dynds /go/bin/do-dyndns

RUN chmod +x /go/bin/do-dyndns
RUN chown -R dyndns:dyndns /go/bin/do-dyndns

USER 10001

ENTRYPOINT [ "/go/bin/do-dyndns" ]