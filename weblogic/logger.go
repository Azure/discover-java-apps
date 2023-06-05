package weblogic

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"strings"
)

type CustomLogger struct {
	logr logr.Logger
}

func GetAzureLogger(ctx context.Context, annotationsMaps ...map[string]string) *CustomLogger {
	log, _ := logr.FromContext(ctx)
	c := &CustomLogger{logr: log}
	if annotationsMaps != nil {
		for _, annotationsMap := range annotationsMaps {
			c.addAzureAnnotations(annotationsMap)
		}
	}
	return c
}

// addAzureAnnotations Adding annotation in custom logger.
func (c *CustomLogger) addAzureAnnotations(annotationsMap map[string]string) {
	for annotationkey, annotationvalue := range annotationsMap {
		// only adding azure annotations
		if strings.HasPrefix(annotationkey, "management.azure.com/") {
			scrubbedKey := strings.TrimPrefix(annotationkey, "management.azure.com/")
			// strcase.ToCamel() and serialize the system data.
			c.logr = c.logr.WithValues(scrubbedKey, annotationvalue)

		} else {
			c.logr = c.logr.WithValues(annotationkey, annotationvalue)
		}
	}
}

func toFields(keysAndValues ...interface{}) map[string]interface{} {
	var m = make(map[string]interface{})
	if len(keysAndValues) == 0 {
		return m
	}

	length := len(keysAndValues)
	for i := 0; i < length; i = i + 2 {
		var key = fmt.Sprintf("%v", keysAndValues[i])
		var val any
		if i+1 < length {
			val = keysAndValues[i+1]
		}
		m[key] = val
	}

	return m
}

func (c *CustomLogger) Info(msg string, keysAndValues ...interface{}) {
	_, l := c.logr.WithCallStackHelper()
	l.V(0).Info(msg, "Fields", toFields(keysAndValues...))
}

func (c *CustomLogger) Debug(msg string, keysAndValues ...interface{}) {
	_, l := c.logr.WithCallStackHelper()
	l.V(1).Info(msg, "Fields", toFields(keysAndValues...))
}

func (c *CustomLogger) Warning(err error, msg string, keysAndValues ...interface{}) {
	if c.logr.GetSink() != nil {
		c.logr.GetSink().WithValues("err", err).Info(-1, msg, "Fields", toFields(keysAndValues...))
	}
}

func (c *CustomLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	_, l := c.logr.WithCallStackHelper()
	l.Error(err, msg, "Fields", toFields(keysAndValues...))
}
