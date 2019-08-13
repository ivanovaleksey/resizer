.PHONY: build
build:
	go build -o bin/app ./cmd

.PHONY: run
run:
	go run ./cmd
