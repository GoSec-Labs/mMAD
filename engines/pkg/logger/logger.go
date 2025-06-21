package logger

import (
	"context"
	"fmt"

	"os"
	"strings"

	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
	"github.com/sirupsen/logrus"
)

type Fields map[string]interface{}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
	InfoWithFields(msg string, fields Fields)
	ErrorWithFields(msg string, fields Fields)
	InfoCtx(ctx context.Context, msg string, fields Fields)
	ErrorCtx(ctx context.Context, msg string, fields Fields)
	WithComponent(component string) Logger
}

type logrusLogger struct {
	logger    *logrus.Logger
	component string
}

func New(c config.LoggingConfig) (Logger, error) {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set formatter
	if strings.ToLower(c.Format) == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false,
		})
	}

	// Set output
	if c.Output != "" && c.Output != "stdout" {
		file, err := os.OpenFile(c.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", c.Output, err)
		}
		log.SetOutput(file)
	}

	return &logrusLogger{
		logger: log,
	}, nil
}

// Basic logging methods
func (l *logrusLogger) Info(msg string) {
	l.logger.WithField("component", l.component).Info(msg)
}

func (l *logrusLogger) Error(msg string) {
	l.logger.WithField("component", l.component).Error(msg)
}

func (l *logrusLogger) Debug(msg string) {
	l.logger.WithField("component", l.component).Debug(msg)
}

func (l *logrusLogger) Warn(msg string) {
	l.logger.WithField("component", l.component).Warn(msg)
}

// Structured logging
func (l *logrusLogger) InfoWithFields(msg string, fields Fields) {
	l.logger.WithFields(logrus.Fields(fields)).WithField("component", l.component).Info(msg)
}

func (l *logrusLogger) ErrorWithFields(msg string, fields Fields) {
	l.logger.WithFields(logrus.Fields(fields)).WithField("component", l.component).Error(msg)
}

// Context-aware logging
func (l *logrusLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
	entry := l.logger.WithContext(ctx)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.WithField("component", l.component).Info(msg)
}

func (l *logrusLogger) ErrorCtx(ctx context.Context, msg string, fields Fields) {
	entry := l.logger.WithContext(ctx)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.WithField("component", l.component).Error(msg)
}

// Component-specific logger
func (l *logrusLogger) WithComponent(component string) Logger {
	return &logrusLogger{
		logger:    l.logger,
		component: component,
	}
}

// Global convenience functions (for backward compatibility)
var defaultLogger Logger

func Init(c config.LoggingConfig) error {
	var err error
	defaultLogger, err = New(c)
	return err
}

func Info(msg string) {
	if defaultLogger != nil {
		defaultLogger.Info(msg)
	}
}

func Error(msg string) {
	if defaultLogger != nil {
		defaultLogger.Error(msg)
	}
}
