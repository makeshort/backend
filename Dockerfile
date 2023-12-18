FROM golang:1.21.5-alpine3.18 AS builder

RUN go version
RUN apk add git

COPY ./ /makeshort-backend
WORKDIR /makeshort-backend

RUN go mod download
RUN go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go


# Lightweight docker container with binary files
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /makeshort-backend/.bin/ ./.bin
COPY --from=builder /makeshort-backend/config/ ./config

CMD ["./.bin/makeshort-backend"]