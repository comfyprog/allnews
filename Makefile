.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: build
build:
	go build -o ./bin/allnews .
