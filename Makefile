.PHONY: build run clean
.DEFAULT_GOAL := build

run:
	./tmp/server_bun

build:
	templ generate
	go build -o ./tmp/server app/server/main.go
	go build -o ./tmp/server_bun app/server_bun/main.go

dbstart:
	turso dev --db-file db-data/local.db --port $TURSO_PORT

dbshell:
	turso db shell http://$TURSO_IP:$TURSO_PORT

clean:
	rm -rf tmp && find . -type f -name '*_templ.go' -delete
