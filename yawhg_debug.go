package yawhg

import (
	"context"
	"fmt"
	"strings"
)

// Debug logs at the debug severity level (staging and development)
func (f *Fields) Debug(msg string) {
	(*f)["severity"] = DebugLevel.String()
	(*f)["msg"] = msg

	structuredWrap(f)
}

// Debugf logs at the debug severity level (staging and development) with a formatting directive
func (f *Fields) Debugf(format string, v ...interface{}) {
	(*f)["severity"] = DebugLevel.String()
	(*f)["msg"] = fmt.Sprintf(format, v...)
	structuredWrap(f)
}

// Debugw is a cumulative logger method for the Fields map that logs at the debug level
func (f *Fields) Debugw(details Fields) {
	for k, v := range details {
		(*f)[k] = v
	}

	(*f)["severity"] = DebugLevel.String()
	structuredWrap(f)
}

// Debug logs a message at the debug severity level
func Debug(v ...interface{}) {
	message := make([]string, len(v))
	for i, value := range v {
		message[i] = fmt.Sprint(value)
	}

	textWrap(context.Background(), strings.Join(message, ", "), DebugLevel.String())
}

// Debugf logs a message with a formatting directive at the debug severity level
func Debugf(format string, v ...interface{}) {
	Debugft(context.Background(), format, v...)
}

// Debugft creates a debug-level log from a string template, and extracts tracing information from context
func Debugft(ctx context.Context, format string, v ...interface{}) {
	textWrap(ctx, fmt.Sprintf(format, v...), DebugLevel.String())
}

// Debugw creates a debug-level log from a map
func Debugw(f Fields) {
	f["severity"] = DebugLevel.String()
	structuredWrap(&f)
}

// DebugWithTracing creates a debug-level log from a map, and extracts tracing information from context
func DebugWithTracing(ctx context.Context, f Fields, errors ...error) {
	addErrors(f, errors)
	f["severity"] = DebugLevel.String()
	ctx = f.addTracing(ctx)
	structuredWrap(&f)
}
