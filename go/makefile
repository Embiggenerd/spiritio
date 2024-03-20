SHELL := /bin/bash

build:
	@CGO_ENABLED=1 go build  -o ./bin/spiritio ./cmd

local:
	CGO_ENABLED=1 go run ./cmd

dev: build
	@./bin/spiritio -go_env dev

prod:build
	@./prod.sh

stop: 
	@./stop.sh
test:
	@go test -v ./...
