swag:
	swag init -g cmd/makeshort-backend/main.go

build:
	go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go

run: swag build
	./.bin/makeshort-backend