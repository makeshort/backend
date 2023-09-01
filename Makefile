swag:
	swag init -g cmd/makeshort-backend/main.go

lint:
	golangci-lint run -D govet -E bodyclose -E contextcheck -E dupl -E goconst

test:
	go test -race ./...

build:
	go build -o ./.bin/makeshort-backend ./cmd/makeshort-backend/main.go

run: swag build
	./.bin/makeshort-backend

docker-build: swag
	docker build -t makeshort-backend:latest .

up: docker-build
	docker compose up -d

down:
	docker compose down