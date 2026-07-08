.PHONY: build dev clean

BINARY_NAME=stonks

build:
	go build -ldflags="-w -s" -o ${BINARY_NAME} cmd/web/main.go

dev:
	go run cmd/web/main.go

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -rf cache/*
	touch cache/.gitkeep
