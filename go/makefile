build:
	@go build -o ./bin/spiritio ./cmd

local:
	go run ./cmd

dev: build
	@./bin/spiritio -go_env dev

prod:build
	@./bin/spiritio -go_env prod

test:
	@go test -v ./...