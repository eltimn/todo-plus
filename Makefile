.PHONY: build run clean
.DEFAULT_GOAL := build

run:
	./tmp/server_bun

build:
	templ generate && go build -o ./tmp/server_bun cmd/server_bun/main.go && go build -o ./tmp/server cmd/server/main.go

clean:
	rm -rf tmp && find . -type f -name '*_templ.go' -delete
