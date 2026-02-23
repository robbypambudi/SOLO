.PHONY: build run test clean tidy

BINARY=solo

build:
	go build -o $(BINARY) ./cmd/solo

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy
