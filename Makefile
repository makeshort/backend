build:
	go build -o ./.bin/makeshort-backend ./cmd/main.go

run:
	./.bin/makeshort-backend

swag:
	swag init -g cmd/main.go