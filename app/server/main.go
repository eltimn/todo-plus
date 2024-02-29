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
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/util"
)

func main() {
	// setup logging
	logLevel, logHandler := logging.Configure(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_HANDLER"))
	slog.Info("Configured logging", slog.String("level", logLevel), slog.String("handler", logHandler))

	// grab some env vars
	listenAddress := util.GetEnv("WEB_LISTEN", ":8989")
	dbUrl := util.GetEnv("DB_URL", "http://127.0.0.1:5000")

	// init libsql db
	database, err := models.OpenDB(dbUrl)
	if err != nil {
		slog.Error("Error connecting to libsql", errs.ErrAttr(err))
		os.Exit(1)
	}

	slog.Info("Connected to libsql", slog.String("url", dbUrl))

	routeEnv := routes.RouteEnv{
		Users: models.NewUserModel(database, 5*time.Second),
	}

	// create router
	router := routes.Routes(&routeEnv)

	// start server
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	server := &http.Server{
		Addr:    listenAddress,
		Handler: router.ServeMux,
	}

	slog.Info("Server starting", "listen", listenAddress)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", errs.ErrAttr(err))
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
		slog.Error("HTTP shutdown error", errs.ErrAttr(err))
		exitCode = 1
	}
	if err := database.Close(); err != nil {
		slog.Error("Error closing database", errs.ErrAttr(err))
		exitCode = 1
	}
	slog.Info("Graceful shutdown complete.")
	os.Exit(exitCode)
}
