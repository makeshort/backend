swag:
	swag init -g cmd/makeshort-backend/main.go

lint:
	golangci-lint run -D govet -E bodyclose -E contextcheck -E dupl -E goconst

build: swag
	go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go

run: build
	./.bin/makeshort-backend