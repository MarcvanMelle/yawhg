package yawhg

import (
	"fmt"
	"strings"
)

// Level type
type Level int32

// enums for error levels
const (
	DebugLevel Level = iota
	InfoLevel
	ErrorLevel
)

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case ErrorLevel:
		return "error"
	case InfoLevel:
		return "info"
	}

	return "unknown"
}

// parseLevel takes a string level and returns the level enum
func parseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "info":
		return InfoLevel, nil
	case "error":
		return ErrorLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid Level: %q", lvl)
}
