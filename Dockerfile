FROM golang:1.27.1-alpine3.24@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

WORKDIR $GOPATH/src/do-dyndns/app/
COPY app/go.mod .
COPY app/go.sum .

RUN go mod download

COPY app/ .

WORKDIR $GOPATH/src/do-dyndns/app/cmd/do-dyndns/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/do-dynds

FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

COPY --from=builder --chmod=0755 /go/bin/do-dynds /go/bin/do-dyndns

USER 10001:10001

ENTRYPOINT [ "/go/bin/do-dyndns" ]
