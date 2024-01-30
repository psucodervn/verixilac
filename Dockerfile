ARG BINARY=verixilac
ARG DIR=/app

FROM golang:1.21-alpine AS builder
ARG BINARY
ARG DIR

RUN apk update && apk add --no-cache git ca-certificates gcc g++

WORKDIR $DIR
COPY go.mod go.sum ./
RUN go mod graph | grep -v '@.*@' | cut -d ' ' -f 2 | xargs go get -v

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $BINARY main.go

FROM alpine
ARG BINARY
ARG DIR

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder $DIR/$BINARY ./app
ENTRYPOINT ["./app"]
