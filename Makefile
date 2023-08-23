swag:
	swag init -g cmd/makeshort-backend/main.go

lint:
	golangci-lint run -D govet -E bodyclose -E contextcheck -E dupl -E goconst

build:
	go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go

run: swag build
	./.bin/makeshort-backend