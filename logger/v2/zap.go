package v2

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Unexported new type so that our context key never collides with another.
// taken from HCLog
type contextKeyType struct{}

// contextKey is the key used for the context to store the logger.
var contextKey = contextKeyType{}

// Option defines a function that allows altering a zap.Config attributes
type Option func(*zap.Config)

// SetFormat allows us to change the format of a logger. Possible values are
// "json" and "console", but always fall back to "console" if an
// undetermined one is entered
func SetFormat(format string) Option {
	encoding := "console"
	return func(c *zap.Config) {
		c.Encoding = encoding
		c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if strings.ToLower(format) == "json" {
			c.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
			c.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
		}
	}
}

// ZapLogger wraps a zap.SugaredLogger and implements the Logger interface in
// this package
type ZapLogger struct {
	*zap.SugaredLogger
}

// Debug ...
func (z *ZapLogger) Debug(msg string, values ...Values) {
	keyAndvalues := []interface{}{}
	for _, value := range values {
		keyAndvalues = append(keyAndvalues, value.toVariadic()...)
	}
	z.Debugw(msg, keyAndvalues...)
}

// Error ...
func (z *ZapLogger) Error(msg string, values ...Values) {
	keyAndvalues := []interface{}{}
	for _, value := range values {
		keyAndvalues = append(keyAndvalues, value.toVariadic()...)
	}
	z.Errorw(msg, keyAndvalues...)
}

// Info ...
func (z *ZapLogger) Info(msg string, values ...Values) {
	keyAndvalues := []interface{}{}
	for _, value := range values {
		keyAndvalues = append(keyAndvalues, value.toVariadic()...)
	}
	z.Infow(msg, keyAndvalues...)
}

// Warn ...
func (z *ZapLogger) Warn(msg string, values ...Values) {
	keyAndvalues := []interface{}{}
	for _, value := range values {
		keyAndvalues = append(keyAndvalues, value.toVariadic()...)
	}
	z.Warnw(msg, keyAndvalues...)
}

// WithValues ...
func (z *ZapLogger) WithValues(values Values) Logger {
	return &ZapLogger{z.With(values.toVariadic()...)}
}

// Close flushes any buffered log entries.
func (z *ZapLogger) Close() {
	if err := z.Sync(); err != nil {
		panic(err)
	}
}

func newZapLogger(setters ...Option) *zap.Logger {
	// Default config
	config := zap.Config{
		Encoding:          "json",
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
		EncoderConfig:     newZapEncoderConfig(),
	}

	for _, setter := range setters {
		setter(&config)
	}

	// We need to skip one caller, since we are going to wrap some functions
	logger, _ := config.Build(zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
	return logger
}

func newZapEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		CallerKey:    "caller",
		TimeKey:      "time",
		LevelKey:     "level",
		MessageKey:   "msg",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.RFC3339TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
}

// NewZapLogger returns a ZapLogger and also allows us to customize some of this
// configuration by using functional options
func NewZapLogger(setters ...Option) Logger {
	return &ZapLogger{newZapLogger(setters...).Sugar()}
}

// FromContext returns a logger from the context. A JSON formatted Zap logger if
// any logger is
func FromContext(ctx context.Context) Logger {
	logger, _ := ctx.Value(contextKey).(Logger)
	if logger == nil {
		return NewZapLogger(SetFormat("json"))
	}

	return logger
}

// WithContext injects a logger into the context that can be retrieved using
// the FromContext function. An optional Values can be passed to add fixed values
// to a new loger.
func WithContext(ctx context.Context, logger Logger, values Values) context.Context {
	if len(values) > 0 {
		logger = logger.WithValues(values)
	}
	return context.WithValue(ctx, contextKey, logger)
}
