package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eltimn/todo-plus/cmd/server/logging"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/routes"
	"eltimn/todo-plus/utils"
)

func main() {
	// grab some env vars
	uri, ok := os.LookupEnv("MONGODB_URI")
	if !ok {
		panic("You must set a 'MONGODB_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	listenAddress, ok := os.LookupEnv("WEB_LISTEN")
	if !ok {
		panic("You must set a 'WEB_LISTEN' environment variable. See\n\t https://golang.org/pkg/net/http/#ListenAndServe")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	logHandler := os.Getenv("LOG_HANDLER")

	// setup logging
	level, handler := logging.Configure(logLevel, logHandler)
	slog.Info("Configured logging", slog.String("level", level), slog.String("handler", handler))

	// init mongodb
	if err := models.InitMongoDB(uri); err != nil {
		slog.Error("Error connecting to MongoDB", utils.ErrAttr(err))
		os.Exit(1)
	}
	slog.Info("Connected to MongoDB", slog.String("uri", uri))

	// add handlers
	routes.Routes()

	// start server
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	server := &http.Server{
		Addr: listenAddress,
	}

	slog.Info("Server starting", "listen", listenAddress)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", utils.ErrAttr(err))
			os.Exit(1)
		}
		slog.Info("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	exitCode := 0
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", utils.ErrAttr(err))
		exitCode = 1
	}
	if err := models.ShutdownMongoDB(); err != nil {
		slog.Error("Error disconnecting from MongoDB", utils.ErrAttr(err))
		exitCode = 1
	}
	slog.Info("Graceful shutdown complete.")
	os.Exit(exitCode)
}
