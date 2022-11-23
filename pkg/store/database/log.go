package database

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
	"gorm.io/gorm/logger"
)

// tracelogger 集成traceid
type tracelogger struct {
	lvl logger.LogLevel
}

func (l *tracelogger) LogMode(lvl logger.LogLevel) logger.Interface {
	newlog := *l
	newlog.lvl = lvl
	return &newlog
}
func (l *tracelogger) Info(ctx context.Context, s string, v ...interface{})  {}
func (l *tracelogger) Warn(ctx context.Context, s string, v ...interface{})  {}
func (l *tracelogger) Error(ctx context.Context, s string, v ...interface{}) {}
func (l *tracelogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.lvl <= logger.Silent {
		return
	}
	log := slog.FromContext(ctx)
	span := trace.SpanFromContext(ctx)
	sql, rows := fc()
	dur := time.Since(begin)
	logattr := []any{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("latency", dur),
		slog.String("traceid", span.SpanContext().TraceID().String()),
	}
	switch {
	case err != nil && l.lvl >= logger.Error:
		log.Error("gorm", err, logattr...)
	case dur >= time.Millisecond*500 && l.lvl >= logger.Warn:
		log.Warn("gorm slow sql", logattr...)
	case l.lvl == logger.Info:
		log.Info("gorm", logattr...)
	}
}
