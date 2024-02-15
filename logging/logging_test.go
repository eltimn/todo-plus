package logging

import (
	"log/slog"
	"testing"
)

func TestLogLevelFromString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want slog.Level
	}{
		{
			name: "Test debug level",
			in:   "debug",
			want: slog.LevelDebug,
		},
		{
			name: "Test info level",
			in:   "info",
			want: slog.LevelInfo,
		},
		{
			name: "Test warn level",
			in:   "warn",
			want: slog.LevelWarn,
		},
		{
			name: "Test error level",
			in:   "error",
			want: slog.LevelError,
		},
		{
			name: "Test default level",
			in:   "unknown",
			want: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logLevelFromString(tt.in); got != tt.want {
				t.Errorf("logLevelFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
