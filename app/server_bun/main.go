package main

import (
	"context"
	"eltimn/todo-plus/app/server_bun/routes"
	"eltimn/todo-plus/logging"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/utils"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
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

	// create the router
	// https://gist.github.com/alexaandru/747f9d7bdfb1fa35140b359bf23fa820 <- syntatic sugar around net/http
	router := bunrouter.New(
		bunrouter.WithNotFoundHandler(routes.NotFoundHandler),
		bunrouter.Use(reqlog.NewMiddleware(reqlog.FromEnv("BUNDEBUG"))),
	)

	// add routes
	routes.Routes(router)

	// setup the server
	httpLn, err := net.Listen("tcp", listenAddress)
	if err != nil {
		slog.Error("Error setting listen address", utils.ErrAttr(err))
		os.Exit(1)
	}

	handler := http.Handler(router)
	// handler = cors.Default().Handler(handler)
	// handler = gzhttp.GzipHandler(handler)

	httpServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}

	slog.Info("Server starting", "listen", listenAddress)

	go func() {
		if err := httpServer.Serve(httpLn); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", utils.ErrAttr(err))
		}
	}()

	slog.Info("Press CTRL+C to exit...")
	slog.Info(waitExitSignal().String())

	// Only wait 20 seconds for http server shutdown
	ctx, shutdownRelease := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownRelease()

	// Graceful shutdown.
	exitCode := 0
	if err := httpServer.Shutdown(ctx); err != nil {
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

func waitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}
