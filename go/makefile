SHELL := /bin/bash

build:
	@CGO_ENABLED=1 go build  -o ./bin/spiritio ./cmd

build-docker:
	sudo docker build -t spirit .

local:
	CGO_ENABLED=1 go run ./cmd

dev: build
	@./bin/spiritio -go_env dev

prod:build
	@./prod.sh

docker: build-docker
	sudo docker run --name spiritio -p 8080:8080 spirit

version=n5.1.2
srcPath=tmp/$(version)/src
patchPath=
platform=

install-ffmpeg:
	rm -rf $(srcPath)
	mkdir -p $(srcPath)
	cd $(srcPath) && git clone https://github.com/FFmpeg/FFmpeg .
	cd $(srcPath) && git checkout $(version)
ifneq "" "$(patchPath)"
	cd $(srcPath) && git apply $(patchPath)
endif
	cd $(srcPath) && ./configure --prefix=.. $(configure)
	cd $(srcPath) && make
	cd $(srcPath) && make install

stop: 
	@./stop.sh
test:
	@go test -v ./...
