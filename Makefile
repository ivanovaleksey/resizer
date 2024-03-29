.PHONY: build
build:
	go build -o bin/app ./cmd

.PHONY: run
run:
	go run ./cmd

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: test-race
test-race:
	go test -race -count=1 ./...

.PHONY: generate
generate:
	go generate ./...
