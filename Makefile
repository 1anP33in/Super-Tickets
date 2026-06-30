.PHONY: run build tidy

run:
	go run ./cmd/bot

build:
	go build -o bin/super-tickets-bot ./cmd/bot

tidy:
	go mod tidy
