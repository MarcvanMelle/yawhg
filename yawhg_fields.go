package yawhg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Fields is a map containing the fields to be logged
type Fields map[string]interface{}

// Copy generates, populates, and returns a new Fields literal with the same key value pairs as the receiver
func (f Fields) Copy() Fields {
	newFields := Fields{}
	for key, value := range f {
		newFields[key] = value
	}
	return newFields
}

func (f *Fields) addBaseFields() {
	(*f)["time"] = time.Now().Format(time.RFC3339Nano)
	(*f)["v"] = appVersion
}

// addTrading extracts tracing information from context and adds it to the log, if available
// N.B. for yawhg to successfully retrieve the `x-request-id` key, users of yawhg must set the key through metadata (as in the test cases)
func (f *Fields) addTracing(ctx context.Context) context.Context {
	copyCtx, requestID := FromContext(ctx)
	(*f)[requestIDKey] = requestID

	return copyCtx
}

func (f *Fields) checkSeverityLevel() (Level, error) {
	severityLevel, ok := (*f)["severity"].(string)
	if !ok {
		return 0, fmt.Errorf("severity level not set for %s", f)
	}

	return parseLevel(severityLevel)
}

func (f *Fields) fire() {
	if err := json.NewEncoder(Destination).Encode(f); err != nil {
		fmt.Printf("logging through yawhg: %s", err)
	}
}

func (f *Fields) resetWrapper() {
	for k := range *f {
		delete(*f, k)
	}
}
