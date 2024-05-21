FROM golang:alpine AS builder
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

WORKDIR $GOPATH/src/do-dyndns/app/
COPY app/go.mod .
COPY app/go.sum .

RUN go mod download

COPY app/ .

WORKDIR $GOPATH/src/do-dyndns/app/cmd/do-dyndns/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/do-dynds

FROM alpine
LABEL org.opencontainers.image.authors="martin@hellstrom.it"

RUN addgroup -S dyndns && adduser -S dyndns -G dyndns

COPY --from=builder /go/bin/do-dynds /go/bin/do-dyndns

RUN chmod +x /go/bin/do-dyndns
RUN chown -R dyndns:dyndns /go/bin/do-dyndns

USER dyndns

ENTRYPOINT [ "/go/bin/do-dyndns" ]