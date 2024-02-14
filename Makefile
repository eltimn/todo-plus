.PHONY: dev build run clean
.DEFAULT_GOAL := dev

rzen:
	go run cmd/rzen/main.go

run:
	go run main.go serve

#--cmd="go run main.go serve"
dev:
	templ generate --watch

devbox:
	devbox shell #--env-file .env.devbox

# dev:
# 	@templ generate --watch --cmd="go run main.go serve"

build:
	templ generate && go build -o ./tmp/server cmd/server/main.go

# run:
#   templ generate --watch --proxy="http://localhost:8090" --cmd="go run main.go serve"

clean:
	rm -rf tmp && rm -rf *_templ.go
