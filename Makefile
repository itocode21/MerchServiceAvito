.PHONY: all build run test-unit test-e2e test-load docker-up docker-down clean

all: build test-unit test-e2e test-load docker-up

build:
	go build -o MerchServiceAvito ./cmd/server/main.go

run:
	go run ./cmd/server/main.go

test-unit:
	go test ./internal/services -v -cover

test-e2e:
	go test -v .

test-load:
	-k6 run load_test.js

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

clean:
	rm -f MerchServiceAvito
	docker-compose rm -f