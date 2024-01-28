package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/constants"
	slogmulti "github.com/samber/slog-multi"
)

const (
	logFatal = slog.Level(13)
)

// Loger is an extension of log.slog that includes fatal
type Logger interface {
	Fatal(msg string)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// NewLoggerService creates and returns a new Logger instance
func NewLoggerService(ctx context.Context, cfg *config.Config) Logger {
	file, _ := os.OpenFile("pkg/logger/"+cfg.LogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.OS_ALL_R)
	stderr := os.Stderr

	slogger := slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(file, &slog.HandlerOptions{}),   // pass to first handler: logstash over tcp
			slog.NewTextHandler(stderr, &slog.HandlerOptions{}), // then to second handler: stderr
		),
	)
	logger := &CustomLogger{Logger: slogger}
	logger.Info("Service Up")
	return logger
}

// CustomLogger implements slog.Handler with custom behavior
type CustomLogger struct {
	*slog.Logger
}

// Fatal logs a message and exits
func (l *CustomLogger) Fatal(msg string) {
	l.Log(nil, logFatal, "msg")
	os.Exit(1)
}

// type NewJSONHandlerWithMetadata struct {
// 	slog.JSONHandler
// }

// func (h *NewJSONHandlerWithMetadata) Handle(ctx context.Context, r slog.Record) error {
// 	// We are adding all of the metadata as a json value on our slog handler
// 	md := utils.ExposeContextMetadata(ctx).ToJSON()
// 	attr := slog.String(string(utils.Metadata_name), md)
// 	r.AddAttrs(attr)

// 	// We check if there is a databa

// 	return h.JSONHandler.Handle(ctx, r)
// }
