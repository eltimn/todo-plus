package logging

import (
	"fmt"
	"log/slog"
	"os"
)

// https://betterstack.com/community/guides/logging/logging-in-go/

func logLevelFromString(in string) slog.Level {
	switch in {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func logHandlerFromString(in string, opts *slog.HandlerOptions) slog.Handler {
	switch in {
	case "json":
		return slog.NewJSONHandler(os.Stdout, opts)
	default:
		return slog.NewTextHandler(os.Stdout, opts)
	}
}

func Configure(logLevel string, logHandler string) (string, string) {
	level := logLevelFromString(logLevel)
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := logHandlerFromString(logHandler, opts)
	logger := slog.New(handler)
	// child := logger.With(
	// 	slog.Group("program_info",
	// 		slog.Int("pid", os.Getpid()),
	// 		slog.String("go_version", runtime.Version()),
	// 	),
	// )

	slog.SetDefault(logger)

	return level.String(), fmt.Sprintf("%T", handler)
}
