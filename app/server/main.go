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

	"eltimn/todo-plus/app/server/routes"
	"eltimn/todo-plus/logging"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/utils"
)

func main() {
	// setup logging
	logLevel, logHandler := logging.Configure(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_HANDLER"))
	slog.Info("Configured logging", slog.String("level", logLevel), slog.String("handler", logHandler))

	// grab some env vars
	mongoUri, ok := os.LookupEnv("MONGODB_URI")
	if !ok {
		mongoUri = "mongodb://localhost:27017/todo"
	}

	listenAddress, ok := os.LookupEnv("WEB_LISTEN")
	if !ok {
		listenAddress = ":8989"
	}

	// init mongodb
	if err := models.InitMongoDB(mongoUri); err != nil {
		slog.Error("Error connecting to MongoDB", utils.ErrAttr(err))
		os.Exit(1)
	}
	slog.Info("Connected to MongoDB", slog.String("uri", mongoUri))

	// create router
	router := routes.Routes()

	// start server
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	server := &http.Server{
		Addr:    listenAddress,
		Handler: router.ServeMux,
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
