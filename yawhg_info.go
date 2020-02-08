package yawhg

import (
	"context"
	"fmt"
	"strings"
)

// Info logs at the info level
func (f *Fields) Info(msg string) {
	(*f)["severity"] = InfoLevel.String()
	(*f)["msg"] = msg
	structuredWrap(f)
}

// Infof logs at the info level with a formatting directive
func (f *Fields) Infof(format string, v ...interface{}) {
	(*f)["severity"] = InfoLevel.String()
	(*f)["msg"] = fmt.Sprintf(format, v...)
	structuredWrap(f)
}

// Infow is a cumulative logger method for the Fields map that logs at the info level
func (f *Fields) Infow(details Fields) {
	for k, v := range details {
		(*f)[k] = v
	}

	(*f)["severity"] = InfoLevel.String()
	structuredWrap(f)
}

// Info logs a message at the info severity level
func Info(v ...interface{}) {
	message := make([]string, len(v))
	for i, value := range v {
		message[i] = fmt.Sprint(value)
	}

	textWrap(context.Background(), strings.Join(message, ", "), InfoLevel.String())
}

// Infof logs a message with a formatting directive at the info severity level
func Infof(format string, v ...interface{}) {
	Infoft(context.Background(), format, v...)
}

// Infoft creates an info-level log from a string template, and extracts tracing information from context
func Infoft(ctx context.Context, format string, v ...interface{}) {
	textWrap(ctx, fmt.Sprintf(format, v...), InfoLevel.String())
}

// Infow creates an info-level log from a map
func Infow(f Fields) {
	f["severity"] = InfoLevel.String()
	structuredWrap(&f)
}

// InfoWithTracing creates an info-level log from a map, and extracts tracing information from context
func InfoWithTracing(ctx context.Context, f Fields, errors ...error) {
	addErrors(f, errors)
	f["severity"] = InfoLevel.String()
	ctx = f.addTracing(ctx)
	structuredWrap(&f)
}
