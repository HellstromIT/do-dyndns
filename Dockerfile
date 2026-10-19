FROM golang:1.27.2-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder
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
