# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.22 AS build-stage

WORKDIR /build

# Install deps
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy sources
COPY *.go ./
COPY logging ./logging
COPY models ./models
COPY pkg ./pkg
COPY routes ./routes
COPY web ./web

# Build
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-plus main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /app

COPY --from=build-stage /todo-plus ./todo-plus
COPY dist/assets ./assets

EXPOSE 8080

USER nonroot:nonroot

CMD ["/app/todo-plus"]

# docker run -p 8989:8989 -e ASSETS_PATH='/app/assets' -e DB_URL=http://192.168.1.43:42069 todo
