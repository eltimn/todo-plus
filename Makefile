.PHONY: build run clean
.DEFAULT_GOAL := build

run:
	go run main.go serve

build:
	templ generate && go build -o ./tmp/server cmd/server/main.go

clean:
	rm -rf tmp && find . -type f -name '*_templ.go' -delete
