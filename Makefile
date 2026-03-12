.PHONY: build run test clean

build:
	go build -o logforwarder main.go

run:
	go run main.go

docker-build:
	docker build -t logforwarder .

docker-run:
	docker run -p 5044:5044 -p 24224:24224 logforwarder

test-up:
	docker-compose up -d

test-down:
	docker-compose down

clean:
	rm -f logforwarder
	docker-compose down -v
