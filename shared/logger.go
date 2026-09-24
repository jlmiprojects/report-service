package utils

import (
	"log/slog"
	"os"
)

type LoggerConfig struct {
	Type  string `mapstructure:"type"`
	Level string `mapstructure:"level"`
}

func SetupLogging(version string, l *LoggerConfig) {

	level := slog.LevelDebug

	if l.Level == "INFO" {
		level = slog.LevelInfo
	} else if l.Level == "WARN" {
		level = slog.LevelWarn
	} else if l.Level == "ERROR" {
		level = slog.LevelError
	} else {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)

	if l.Type == "json" {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("version", version)
		slog.SetDefault(logger)
	} else {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("version", version)
		slog.SetDefault(logger)
	}

}
