package logger

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/constants"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/google/uuid"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/negroni"
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
	LoggingMW(next http.Handler) http.Handler
	LogAPIRequest(id, ip, path, port, method string, timeRecieved time.Time, nanoSeconds int64, statusCode int)
	LogRequestError(requestID, errorMessage string, statusCode int)
	// logMessage(ctx context.Context direction, message, data string)
	LogMessageSent(ctx context.Context, message *types.WebsocketMessage)
	LogWorkOrderReceived(ctx context.Context, workOrder *types.WorkOrder)
}

// replaceAttr masks data from requests and metadata from context
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "data" || a.Key == "metadata" {
		a = slog.Attr{}
	}
	return a
}

// NewLoggerService creates and returns a new Logger instance
func NewLoggerService(ctx context.Context, cfg *config.Config) Logger {
	file, err := os.OpenFile("pkg/logger/"+cfg.LogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.OS_ALL_RW)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	slogger := slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(file, &slog.HandlerOptions{
				AddSource: true,
			}),
			NewPrettyHandler(&slog.HandlerOptions{
				Level:       slog.LevelInfo,
				AddSource:   true,
				ReplaceAttr: replaceAttr,
			}),
		),
	)
	logger := &CustomLogger{Logger: slogger}
	logger.Info("logging service Up")
	return logger
}

// CustomLogger implements slog.Handler with custom behavior
type CustomLogger struct {
	*slog.Logger
}

// Fatal logs a message and exits
func (l *CustomLogger) Fatal(msg string) {
	l.Log(context.TODO(), logFatal, "msg")
	os.Exit(1)
}

func (l *CustomLogger) LoggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(utils.WithMetadata(context.Background()))
		defer cancel()

		method := r.Method
		path := r.URL.EscapedPath()
		ip, port, _ := net.SplitHostPort(r.RemoteAddr)
		lrw := negroni.NewResponseWriter(w)
		newUUID := uuid.New()

		utils.ExposeContextMetadata(ctx).Set("requestID", newUUID.String())

		next.ServeHTTP(w, r.WithContext(ctx))

		statusCode := lrw.Status()

		defer func(begin time.Time) {
			tookMs := time.Since(begin).Nanoseconds()
			l.LogAPIRequest(newUUID.String(), ip, path, port, method, time.Now(), tookMs, statusCode)
		}(time.Now())
	})
}

func (l *CustomLogger) LogAPIRequest(id, ip, path, port, method string, timeRecieved time.Time, nanoSeconds int64, statusCode int) {
	level := slog.LevelInfo
	if statusCode >= 400 {
		level = slog.LevelError
	}

	l.Log(context.TODO(), level, "API Request",
		slog.String("requestID", id),
		slog.Int("statusCode", statusCode),
		slog.String("ip", ip),
		slog.String("path", path),
		slog.String("port", port),
		slog.String("method", method),
		slog.Int64("nanoSeconds", nanoSeconds),
		slog.Time("timeReceived", timeRecieved),
	)
}

func (l *CustomLogger) LogRequestError(requestID, errorMessage string, statusCode int) {
	l.Log(
		context.TODO(),
		slog.LevelError,
		"Request Error",
		slog.String("requestID", requestID),
		slog.String("errorMessage", errorMessage),
		slog.Int("statusCode", statusCode),
	)
}

func (l *CustomLogger) LogMessageSent(ctx context.Context, message *types.WebsocketMessage) {
	d, err := json.Marshal(message.Data)
	if err != nil {
		l.Error(err.Error())
		return
	}

	l.logMessage(ctx, "Sent", message.Type, string(d))
}

func (l *CustomLogger) LogWorkOrderReceived(ctx context.Context, workOrder *types.WorkOrder) {

	d, err := json.Marshal(workOrder.Details)
	if err != nil {
		l.Error(err.Error())
		return
	}

	l.logMessage(ctx, "Received", workOrder.Order, string(d))
}
func (l *CustomLogger) logMessage(ctx context.Context, direction, messageType, data string) {
	metadata := utils.ExposeContextMetadata(ctx)
	metadataJSON := metadata.ToJSON()
	requestID, _ := metadata.Get("requestID")

	l.Log(
		context.TODO(),
		slog.LevelInfo,
		"Event Message "+direction,
		slog.String("requestID", requestID.(string)),
		slog.String("type", messageType),
		slog.String("data", data),
		slog.String("metadata", metadataJSON),
	)
}
