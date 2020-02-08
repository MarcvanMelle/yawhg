package yawhg

import (
	"context"
	"fmt"
	"strings"
)

// Error logs at the error level
func (f *Fields) Error(msg string) {
	(*f)["severity"] = ErrorLevel.String()
	(*f)["msg"] = msg
	structuredWrap(f)
}

// Errorf logs at the error level with a formatting directive
func (f *Fields) Errorf(format string, v ...interface{}) {
	(*f)["severity"] = ErrorLevel.String()
	(*f)["msg"] = fmt.Sprintf(format, v...)
	structuredWrap(f)
}

// Errorw is a cumulative logger method for the Fields map that logs at the error level
func (f *Fields) Errorw(details Fields) {
	for k, v := range details {
		(*f)[k] = v
	}

	(*f)["severity"] = ErrorLevel.String()
	structuredWrap(f)
}

// Error logs a message at the error severity level
func Error(v ...interface{}) {
	message := make([]string, len(v))
	for i, value := range v {
		message[i] = fmt.Sprint(value)
	}

	textWrap(context.Background(), strings.Join(message, ", "), ErrorLevel.String())
}

// Errorf logs a message with a formatting directive at the error severity level
func Errorf(format string, v ...interface{}) {
	Errorft(context.Background(), format, v...)
}

// Errorft creates an error-level log from a string template, and extracts tracing information from context
func Errorft(ctx context.Context, format string, v ...interface{}) {
	textWrap(ctx, fmt.Sprintf(format, v...), ErrorLevel.String())
}

// Errorw creates an error-level log from a map
func Errorw(f Fields) {
	f["severity"] = ErrorLevel.String()
	structuredWrap(&f)
}

// ErrorWithTracing creates an error-level log from a map, and extracts tracing information from context
func ErrorWithTracing(ctx context.Context, f Fields, errors ...error) {
	addErrors(f, errors)
	f["severity"] = ErrorLevel.String()
	ctx = f.addTracing(ctx)
	structuredWrap(&f)
}
