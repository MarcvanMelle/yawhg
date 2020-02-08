// Package yawhg is a structured logger
package yawhg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

const RequestIDHeader string = "x-request-id"
const requestIDKey string = "request_id"

var Destination io.Writer

var appVersion string
var appLogLevel Level

// fieldsPool caches allocated but unused items for later reuse,
// relieving pressure on the garbage collector.
var fieldsPool *sync.Pool

// Options is a struct containing initialization options for yawhg
// Disabled controls whether or not the logs will be output to os.Stdout or disposed (useful for test environments)
// AppVersion is the version of the current application.  It will be attached to all logs for troubleshooting purposes.
type Options struct {
	AppVersion string
	Enabled    bool
	LogLevel   string
}

// ConfigYawhg overrides the default yawgh initialization with custom options
func ConfigYawhg(options Options) {
	if !options.Enabled {
		Destination = ioutil.Discard
	}

	appVersion = options.AppVersion

	switch options.LogLevel {
	case "DebugLevel":
		appLogLevel = DebugLevel
	case "InfoLevel":
		appLogLevel = InfoLevel
	case "ErrorLevel":
		appLogLevel = ErrorLevel
	default:
		appLogLevel = InfoLevel
	}
}

// NewLogger returns a map used for cumulative logging
func NewLogger() Fields {
	return make(Fields)
}

// WithFields preserves the signature of our previous logging package and returns a Fields map
// The log will be serialized and written when a level is called, e.g. yawhg.WithFields({}).Info("message")
func WithFields(details Fields, errors ...error) *Fields {
	addErrors(details, errors)
	// make a copy of the map values to prevent a data race during concurrent calls
	data := details.Copy()

	return &data
}

// WithTracing behaved like WithFields, but in addition, will extract tracing information from the supplied context struct and add it to the fields map to be logged
// The log will be serialized and written when a level is called, e.g. yawhg.WithFields({}).Info("message")
func WithTracing(ctx context.Context, details Fields, errors ...error) *Fields {
	addErrors(details, errors)
	// make a copy of the map values to prevent a data race during concurrent calls
	data := details.Copy()
	ctx = data.addTracing(ctx)

	return &data
}

func init() {
	Destination = os.Stdout
	appLogLevel = InfoLevel

	fieldsPool = &sync.Pool{
		New: func() interface{} {
			return &Fields{}
		},
	}
}

func structuredWrap(msgMap *Fields) {
	msgMap.addBaseFields()

	messageLevel, err := msgMap.checkSeverityLevel()
	if err != nil {
		fmt.Printf("checking log message severity level: %v", err)
		messageLevel = InfoLevel // default to InfoLevel in case of error
	}

	// only write the log if the severity level rises to the specified threshold
	if messageLevel >= appLogLevel {
		msgMap.fire()
	}
}

func textWrap(ctx context.Context, msg string, level string) {
	data := fieldsPool.Get().(*Fields)
	defer fieldsPool.Put(data)
	defer data.resetWrapper() // defers are executed as LIFO per https://blog.golang.org/defer-panic-and-recover
	(*data)["severity"] = level
	(*data)["msg"] = msg

	data.addBaseFields()
	ctx = data.addTracing(ctx)

	messageLevel, err := data.checkSeverityLevel()
	if err != nil {
		fmt.Printf("checking log message severity level: %v", err)
		messageLevel = InfoLevel // default to InfoLevel in case of error
	}

	// only write the log if the severity level rises to the specified threshold
	if messageLevel >= appLogLevel {
		data.fire()
	}
}

// add a concatenation of non-nil errors to the "Error" field
func addErrors(f Fields, errors []error) {
	if len(errors) > 0 {
		errStrings := make([]string, 0, len(errors))
		for _, err := range errors {
			if err == nil {
				continue
			}
			errStrings = append(errStrings, err.Error())
		}

		if len(errStrings) > 0 {
			f["Error"] = strings.Join(errStrings, ", ")
		}
	}
}
