FROM golang:1.20.4-alpine3.18 AS builder

RUN go version
RUN apk add git

COPY ./ /makeshort-backend
WORKDIR /makeshort-backend

RUN go mod download && go get -u ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go


# Lightweight docker container with binary files
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=0 /makeshort-backend/.bin/ .
COPY --from=0 /makeshort-backend/config/ .

CMD ["./makeshort-backend"]