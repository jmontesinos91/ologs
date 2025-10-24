// Package logger provides a context to logger for regular - text - logging or
// structured - json - logging.
package logger

import (
	"bytes"
	"log/syslog"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

// Context holds key value pairs to add context to logging
type Context map[string]interface{}

// Format defines how logs are formatted
type Format string

const (
	// TextFormat logs are formatted as text
	TextFormat Format = "TEXT"
	// JSONFormat logs are formatted as json
	JSONFormat Format = "JSON"
)

// ContextLogger provides context for logrus logger
type ContextLogger struct {
	logger  *logrus.Logger
	buf     *bytes.Buffer
	context Context
}

// NewContextLogger creates a new context logger value
func NewContextLogger(application, logLevel string, format Format) *ContextLogger {
	logger := logrus.Logger{
		Formatter: &logrus.TextFormatter{},
		Hooks:     logrus.LevelHooks{},
		Out:       os.Stdout,
	}

	context := Context{
		"application": application,
	}

	if host, err := os.Hostname(); err == nil {
		context["hostname"] = host
	}

	contextLogger := &ContextLogger{
		logger:  &logger,
		context: context,
		buf:     new(bytes.Buffer),
	}

	contextLogger.SetLogFormat(format)
	contextLogger.SetLogLevel(logLevel)

	return contextLogger
}

// WithContext provide context for log entries
func (l ContextLogger) WithContext(level logrus.Level, caller, entry string, context Context, err error) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method": caller,
		},
	)

	if err != nil {
		fields["error"] = err.Error()
	}

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, entry)

	if level == logrus.ErrorLevel || level == logrus.FatalLevel || level == logrus.PanicLevel {
		CaptureError(l.context["application"].(string), fields, caller, entry, err)
	}
}

// Log provide method name context for log entries
func (l ContextLogger) Log(level logrus.Level, caller, entry string) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method": caller,
		},
	)

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, entry)
}

// Error provide method name and error context for log entries
func (l ContextLogger) Error(level logrus.Level, caller, entry string, err error) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method": caller,
		},
	)

	if err != nil {
		fields["error"] = err.Error()
	}

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, entry)

	if level == logrus.ErrorLevel || level == logrus.FatalLevel || level == logrus.PanicLevel {
		CaptureError(l.context["application"].(string), fields, caller, entry, err)
	}
}

// Default provide method name context for log entries
func (l ContextLogger) Default(level logrus.Level, caller string, entry string) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method": caller,
		},
	)

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, entry)
}

// AddSyslogHook add syslog hook to an existing context logger
func (l *ContextLogger) AddSyslogHook(host, port string) {
	hook, err := lSyslog.NewSyslogHook(
		"udp",
		host+":"+port,
		syslog.LOG_DEBUG,
		"",
	)
	if err == nil {
		l.logger.Hooks.Add(hook)
	} else {
		l.logger.Printf("Could not hook to syslog, err %s", err)
	}
}

// SetLogFormat sets the log format for logrus logger
func (l ContextLogger) SetLogFormat(format Format) {
	switch format {
	case JSONFormat:
		l.logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		l.logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
			FullTimestamp:   true,
		})
	}
}

// SetLogLevel sets the log level for logrus logger
func (l ContextLogger) SetLogLevel(logLevel string) {
	level := getLogrusLevel(logLevel)
	l.logger.SetLevel(level)
}

// SetSentryDsn sets the sentry dsn connection to send error level logs
func (l ContextLogger) SetSentryDsn(dsn string, environment string, debug bool) {
	ctx := l.logger.WithFields(logrus.Fields(map[string]interface{}{}))

	err := initSentry(dsn, environment, debug)
	if err != nil {
		l.log(logrus.WarnLevel, ctx, "ContextLogger will not send error logs to sentry because DSN failed to connect: "+err.Error())
	} else {
		l.log(logrus.InfoLevel, ctx, "ContextLogger will send error logs to sentry")
	}
}

// InvalidParameter provide method name and parameter context for log entries
func (l ContextLogger) InvalidParameter(level logrus.Level, caller, parameter string, err error) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method":    caller,
			"parameter": parameter,
		},
	)

	if err != nil {
		fields["error"] = err.Error()
	}

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, "invalid parameter")
}

// InvalidRequestBody provide method name and parameter context for log
// entries
func (l ContextLogger) InvalidRequestBody(level logrus.Level, caller string, err error) {
	fields := l.prepareContext(
		l.context,
		logrus.Fields{
			"method": caller,
		},
	)

	if err != nil {
		fields["error"] = err.Error()
	}

	ctx := l.logger.WithFields(
		logrus.Fields(fields),
	)

	l.log(level, ctx, "invalid request body")
}

// Output it's to get the logger as string
func (l ContextLogger) Output() string {
	return l.buf.String()
}

/******************************************************************************/
/* AUXILIARY FUNCTIONS                                                        */
/******************************************************************************/

func getLogrusLevel(logLevel string) logrus.Level {
	var level logrus.Level
	switch strings.ToLower(logLevel) {
	case "trace":
		level = logrus.TraceLevel
	case "debug":
		level = logrus.DebugLevel
	case "warning":
		level = logrus.WarnLevel
	case "info":
		level = logrus.InfoLevel
	default:
		level = logrus.InfoLevel
	}
	return level
}

func (l ContextLogger) prepareContext(context Context, customFields logrus.Fields) logrus.Fields {
	fields := logrus.Fields{}
	l.logger.Out = l.buf
	l.logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	for k, v := range customFields {
		fields[k] = v
	}

	for k, v := range context {
		fields[k] = v
	}

	return fields
}

func (l ContextLogger) log(level logrus.Level, context *logrus.Entry, entry string) {
	switch level {
	case logrus.DebugLevel:
		context.Debug(entry)
	case logrus.ErrorLevel:
		context.Error(entry)
	case logrus.FatalLevel:
		context.Fatal(entry)
	case logrus.PanicLevel:
		context.Panic(entry)
	case logrus.TraceLevel:
		context.Trace(entry)
	case logrus.WarnLevel:
		context.Warn(entry)
	default:
		context.Info(entry)
	}
}
