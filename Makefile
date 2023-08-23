build:
	go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go

run: build
	./.bin/makeshort-backend

swag:
	swag init -g cmd/main.go