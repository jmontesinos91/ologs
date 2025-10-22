package logger

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

// InitSentry sets up the connection to Sentry for error tracking and monitoring.
func initSentry(dsn string, environment string, debug bool) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            debug,
		Environment:      environment,
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
		EnableTracing:    true,
		SendDefaultPII:   true,
	})
}

// CaptureError sends an error to Sentry if one is present
func CaptureError(serviceName string, fields logrus.Fields, caller, entry string, err error) {

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		for k, v := range fields {
			scope.SetContext("log", map[string]interface{}{
				k: v,
			})
		}
		scope.SetTag("method", caller)
		scope.SetTag("service-name", serviceName)
	})

	if err != nil {
		sentry.CaptureException(err)
	}

	defer sentry.Flush(2 * time.Second)

	sentry.CaptureMessage(entry)
}
