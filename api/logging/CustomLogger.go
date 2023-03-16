package logging

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CustomLogger struct {
	logr logr.Logger
}

func GetAzureLogger(ctx context.Context, annotationsMaps ...map[string]string) *CustomLogger {
	c := &CustomLogger{log.FromContext(ctx)}
	return c
}

func (c *CustomLogger) Info(msg string, keysAndValues ...interface{}) {
	c.logr.Info(msg, keysAndValues...)
}

func (c *CustomLogger) Debug(msg string, keysAndValues ...interface{}) {
	c.logr.Info(msg, keysAndValues...)
}

func (c *CustomLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	c.logr.Error(err, msg, keysAndValues...)
}

func (c *CustomLogger) IntoContext(ctx context.Context) context.Context {
	return log.IntoContext(ctx, c.logr)
}
