package yawhg_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/MarcvanMelle/yawhg"
)

type fieldsLogTestCase struct {
	name           string
	errorA         error
	errorB         error
	data           yawhg.Fields
	context        context.Context
	message        string
	style          string
	messageLevel   yawhg.Level
	appLogLevel    yawhg.Level
	template       string
	expectedResult []string
}

var fieldsLogTestCases = []fieldsLogTestCase{
	fieldsLogTestCase{
		name: "info field with string value",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "info message",
		style:          "WithFields",
		messageLevel:   yawhg.InfoLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name: "info field with int value",
		data: yawhg.Fields{
			"Test": 1,
		},
		message:        "info message",
		style:          "WithFields",
		messageLevel:   yawhg.InfoLevel,
		expectedResult: []string{`"Test":1`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name:    "info field with tracing (request_id)",
		context: metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		data: yawhg.Fields{
			"Test": 1,
		},
		message:        "info message",
		style:          "fields WithTracing",
		messageLevel:   yawhg.InfoLevel,
		expectedResult: []string{`"Test":1`, `"request_id":"fake-request-id"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name: "debug field with string value",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "debug message",
		style:          "WithFields",
		messageLevel:   yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"debug"`},
	},
	fieldsLogTestCase{
		name: "debug field with int value",
		data: yawhg.Fields{
			"Test": 1,
		},
		message:        "debug message",
		style:          "WithFields",
		messageLevel:   yawhg.DebugLevel,
		expectedResult: []string{`"Test":1`, `"severity":"debug"`},
	},
	fieldsLogTestCase{
		name: "for multiple fields, do not log debug message when info level logging specified",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "debug message",
		style:          "WithFields",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{},
	},
	fieldsLogTestCase{
		name: "error field with string value",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "error message",
		style:          "WithFields",
		messageLevel:   yawhg.ErrorLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"error"`},
	},
	fieldsLogTestCase{
		name: "error field with int value",
		data: yawhg.Fields{
			"Test": 1,
		},
		message:        "error message",
		style:          "WithFields",
		messageLevel:   yawhg.ErrorLevel,
		expectedResult: []string{`"Test":1`, `"severity":"error"`},
	},
	fieldsLogTestCase{
		name: "for multiple fields, do not log error message when info level logging specified",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "error message",
		style:          "WithFields",
		messageLevel:   yawhg.ErrorLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{},
	},
	fieldsLogTestCase{
		name: "for multiple fields, log info message when debug level logging specified",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "info message",
		style:          "WithFields",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name: "for multiple fields, log info message with formatting directive",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		message:        "template info message",
		template:       "Sending: %v",
		style:          "WithFields template",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{`"Test":"Foo"`, `"msg":"Sending: template info message"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name: "for multiple fields, log at info level when using .Infow",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		style:          "wrapper",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name: "for multiple fields, log at debug level when using .Debugw",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		style:          "wrapper",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"severity":"debug"`},
	},
	fieldsLogTestCase{
		name: "log tracing information with DebugWithTracing",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "withTrace",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"debug"`},
	},
	fieldsLogTestCase{
		name: "log tracing information with InfoWithTracing",
		data: yawhg.Fields{
			"Test": "Foo",
		},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "withTrace",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"info"`},
	},
	fieldsLogTestCase{
		name:   "info log error information with WithTracing",
		errorA: fmt.Errorf("foop"),
		errorB: fmt.Errorf("blarg"),
		data: yawhg.Fields{
			"Test": "Foo",
		},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "fields WithTracing and Errors",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"info"`, `"Error":"foop, blarg"`},
	},
	fieldsLogTestCase{
		name:   "info log error information with WithTracing",
		errorA: fmt.Errorf("foop"),
		errorB: fmt.Errorf("blarg"),
		data: yawhg.Fields{
			"Test": "Foo",
		},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "fields WithTracing and Errors",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"debug"`, `"Error":"foop, blarg"`},
	},
	fieldsLogTestCase{
		name:   "info log WithTracing exclude nil errors",
		errorA: nil,
		errorB: nil,
		data: yawhg.Fields{
			"Test": "Foo",
		},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "fields WithTracing and Errors",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"debug"`},
	},
	fieldsLogTestCase{
		name:           "copy fields",
		data:           yawhg.Fields{"Test": "Foo"},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		style:          "copy",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"Test":"Foo"`, `"request_id":"fake-request-id"`, `"severity":"debug"`},
	},
}

func TestFieldsLog(t *testing.T) {
	for _, testCase := range fieldsLogTestCases {
		yawhg.ConfigYawhg(yawhg.Options{
			Enabled:    false,
			AppVersion: "test",
			LogLevel:   strings.Title(testCase.messageLevel.String()) + "Level",
		})

		// Pipe creates a synchronous in-memory pipe. It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
		r, w := io.Pipe()
		actualResult := new(bytes.Buffer)
		previousDestination := yawhg.Destination
		yawhg.Destination = w // send log output to the pipe instead of os.Stdout

		defer func() {
			yawhg.Destination = previousDestination
		}()

		t.Run(testCase.name, func(t *testing.T) {
			go func() { // each Write to the PipeWriter blocks until one or more Reads from the PipeReader fully consume the written data
				switch testCase.style {
				case "WithFields":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.WithFields(testCase.data).Debug(testCase.message)
					case yawhg.ErrorLevel:
						yawhg.WithFields(testCase.data).Error(testCase.message)
					case yawhg.InfoLevel:
						yawhg.WithFields(testCase.data).Info(testCase.message)
					}
				case "WithFields template":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.WithFields(testCase.data).Debugf(testCase.template, testCase.message)
					case yawhg.ErrorLevel:
						yawhg.WithFields(testCase.data).Errorf(testCase.template, testCase.message)
					case yawhg.InfoLevel:
						yawhg.WithFields(testCase.data).Infof(testCase.template, testCase.message)
					}
				case "wrapper":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.Debugw(testCase.data)
					case yawhg.ErrorLevel:
						yawhg.Errorw(testCase.data)
					case yawhg.InfoLevel:
						yawhg.Infow(testCase.data)
					}
				case "fields WithTracing":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.WithTracing(testCase.context, testCase.data).Debug(testCase.message)
					case yawhg.ErrorLevel:
						yawhg.WithTracing(testCase.context, testCase.data).Error(testCase.message)
					case yawhg.InfoLevel:
						yawhg.WithTracing(testCase.context, testCase.data).Info(testCase.message)
					}
				case "withTrace":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.DebugWithTracing(testCase.context, testCase.data)
					case yawhg.ErrorLevel:
						yawhg.ErrorWithTracing(testCase.context, testCase.data)
					case yawhg.InfoLevel:
						yawhg.InfoWithTracing(testCase.context, testCase.data)
					}
				case "fields WithTracing and Errors":
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.WithTracing(testCase.context, testCase.data, testCase.errorA, testCase.errorB).Debug(testCase.message)
					case yawhg.ErrorLevel:
						yawhg.WithTracing(testCase.context, testCase.data, testCase.errorA, testCase.errorB).Error(testCase.message)
					case yawhg.InfoLevel:
						yawhg.WithTracing(testCase.context, testCase.data, testCase.errorA, testCase.errorB).Info(testCase.message)
					}
				case "copy":
					newFields := testCase.data.Copy()
					yawhg.DebugWithTracing(testCase.context, newFields)
				}
				w.Close()
			}()

			actualResult.ReadFrom(r) // blocks until data is put into the pipe
			for _, result := range testCase.expectedResult {
				if !strings.Contains(actualResult.String(), string(result)) {
					t.Fatalf("expected %v to contain %v", actualResult, result)
				}
			}
		})
	}
}

type stringLogTestCase struct {
	name           string
	data           string
	context        context.Context
	message        string
	style          string
	messageLevel   yawhg.Level
	appLogLevel    yawhg.Level
	template       string
	expectedResult []string
}

var stringLogTestCases = []stringLogTestCase{
	stringLogTestCase{
		name:           "Debug",
		data:           "Debug Test",
		messageLevel:   yawhg.DebugLevel,
		expectedResult: []string{`"msg":"Debug Test"`, `"severity":"debug"`},
	},
	stringLogTestCase{
		name:           "Debugf",
		data:           "Debugf Test",
		template:       "Sending: %v",
		messageLevel:   yawhg.DebugLevel,
		expectedResult: []string{`"msg":"Sending: Debugf Test"`, `"severity":"debug"`},
	},
	stringLogTestCase{
		name:           "Info",
		data:           "Info Test",
		messageLevel:   yawhg.InfoLevel,
		expectedResult: []string{`"msg":"Info Test"`, `"severity":"info"`},
	},
	stringLogTestCase{
		name:           "Infof",
		data:           "Infof Test",
		messageLevel:   yawhg.InfoLevel,
		template:       "Sending: %v",
		expectedResult: []string{`"msg":"Sending: Infof Test"`, `"severity":"info"`},
	},
	stringLogTestCase{
		name:           "for simple message, do not log debug message when info level logging specified",
		data:           "debug message",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{},
	},
	stringLogTestCase{
		name:           "for simple message, log info message when debug level logging specified",
		data:           "info log",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"msg":"info log"`, `"severity":"info"`},
	},
}

func TestSimpleStringLog(t *testing.T) {
	for _, testCase := range stringLogTestCases {
		yawhg.ConfigYawhg(yawhg.Options{
			Enabled:    false,
			AppVersion: "test",
			LogLevel:   strings.Title(testCase.messageLevel.String()) + "Level",
		})

		// Pipe creates a synchronous in-memory pipe. It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
		r, w := io.Pipe()
		actualResult := new(bytes.Buffer)
		previousDestination := yawhg.Destination
		yawhg.Destination = w // send log output to the pipe instead of os.Stdout

		defer func() {
			yawhg.Destination = previousDestination
		}()

		t.Run(testCase.name, func(t *testing.T) {
			go func() { // each Write to the PipeWriter blocks until one or more Reads from the PipeReader fully consume the written data
				if testCase.template != "" {
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.Debugf(testCase.template, testCase.data)
					case yawhg.InfoLevel:
						yawhg.Infof(testCase.template, testCase.data)
					}
				} else {
					switch testCase.messageLevel {
					case yawhg.DebugLevel:
						yawhg.Debug(testCase.data)
					case yawhg.InfoLevel:
						yawhg.Info(testCase.data)
					}
				}

				w.Close()
			}()

			actualResult.ReadFrom(r) // blocks until data is put into the pipe
			for _, result := range testCase.expectedResult {
				if !strings.Contains(actualResult.String(), string(result)) {
					t.Fatalf("expected %v to contain %v", actualResult, result)
				}
			}
		})
	}
}

type multipleStringLogTestCase struct {
	name           string
	data           []string
	context        context.Context
	message        string
	messageLevel   yawhg.Level
	appLogLevel    yawhg.Level
	template       string
	expectedResult []string
}

var multipleStringLogTestCases = []multipleStringLogTestCase{
	multipleStringLogTestCase{
		name:           "Debugf with multiple arguments",
		data:           []string{"Debugf Test", "foo"},
		messageLevel:   yawhg.DebugLevel,
		template:       "Sending: %v, Result: %v",
		expectedResult: []string{`"msg":"Sending: Debugf Test, Result: foo"`, `"severity":"debug"`},
	},
	multipleStringLogTestCase{
		name:           "Infof with multiple arguments",
		data:           []string{"Infof Test", "foo"},
		messageLevel:   yawhg.InfoLevel,
		template:       "Sending: %v, Result: %v",
		expectedResult: []string{`"msg":"Sending: Infof Test, Result: foo"`, `"severity":"info"`},
	},
}

func TestMultipleStringLog(t *testing.T) {
	for _, testCase := range multipleStringLogTestCases {
		yawhg.ConfigYawhg(yawhg.Options{
			Enabled:    false,
			AppVersion: "test",
			LogLevel:   strings.Title(testCase.messageLevel.String()) + "Level",
		})

		// Pipe creates a synchronous in-memory pipe. It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
		r, w := io.Pipe()
		actualResult := new(bytes.Buffer)
		previousDestination := yawhg.Destination
		yawhg.Destination = w // send log output to the pipe instead of os.Stdout

		defer func() {
			yawhg.Destination = previousDestination
		}()

		t.Run(testCase.name, func(t *testing.T) {
			go func() { // each Write to the PipeWriter blocks until one or more Reads from the PipeReader fully consume the written data
				switch testCase.messageLevel {
				case yawhg.DebugLevel:
					yawhg.Debugf(testCase.template, testCase.data[0], testCase.data[1])
				case yawhg.InfoLevel:
					yawhg.Infof(testCase.template, testCase.data[0], testCase.data[1])
				}

				w.Close()
			}()

			actualResult.ReadFrom(r) // blocks until data is put into the pipe
			for _, result := range testCase.expectedResult {
				if !strings.Contains(actualResult.String(), string(result)) {
					t.Fatalf("expected %v to contain %v", actualResult, result)
				}
			}
		})
	}
}

type multipleStringTracingLogTestCase struct {
	name           string
	data           []string
	context        context.Context
	message        string
	messageLevel   yawhg.Level
	appLogLevel    yawhg.Level
	template       string
	expectedResult []string
}

var multipleStringTracingLogTestCases = []multipleStringTracingLogTestCase{
	multipleStringTracingLogTestCase{
		name:           "log debug template message with tracing information",
		data:           []string{"Debugft test", "foo"},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		template:       "Sending: %v, Result: %v",
		messageLevel:   yawhg.DebugLevel,
		appLogLevel:    yawhg.DebugLevel,
		expectedResult: []string{`"msg":"Sending: Debugft test, Result: foo"`, `"request_id":"fake-request-id"`, `"severity":"debug"`},
	},
	multipleStringTracingLogTestCase{
		name:           "log info template message with tracing information",
		data:           []string{"Infoft test", "foo"},
		context:        metadata.AppendToOutgoingContext(context.Background(), yawhg.RequestIDHeader, "fake-request-id"),
		template:       "Sending: %v, Result: %v",
		messageLevel:   yawhg.InfoLevel,
		appLogLevel:    yawhg.InfoLevel,
		expectedResult: []string{`"msg":"Sending: Infoft test, Result: foo"`, `"request_id":"fake-request-id"`, `"severity":"info"`},
	},
}

func TestMultipleStringTracingLog(t *testing.T) {
	for _, testCase := range multipleStringTracingLogTestCases {
		yawhg.ConfigYawhg(yawhg.Options{
			Enabled:    false,
			AppVersion: "test",
			LogLevel:   strings.Title(testCase.messageLevel.String()) + "Level",
		})

		// Pipe creates a synchronous in-memory pipe. It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
		r, w := io.Pipe()
		actualResult := new(bytes.Buffer)
		previousDestination := yawhg.Destination
		yawhg.Destination = w // send log output to the pipe instead of os.Stdout

		defer func() {
			yawhg.Destination = previousDestination
		}()

		t.Run(testCase.name, func(t *testing.T) {
			go func() { // each Write to the PipeWriter blocks until one or more Reads from the PipeReader fully consume the written data
				switch testCase.messageLevel {
				case yawhg.DebugLevel:
					yawhg.Debugft(testCase.context, testCase.template, testCase.data[0], testCase.data[1])
				case yawhg.InfoLevel:
					yawhg.Infoft(testCase.context, testCase.template, testCase.data[0], testCase.data[1])
				}

				w.Close()
			}()

			actualResult.ReadFrom(r) // blocks until data is put into the pipe
			for _, result := range testCase.expectedResult {
				if !strings.Contains(actualResult.String(), string(result)) {
					t.Fatalf("expected %v to contain %v", actualResult, result)
				}
			}
		})
	}
}

func TestCumulativeLogger(t *testing.T) {
	want := []string{`"myField":"foo"`, `"additionalField":"bar"`, `"severity":"info"`}

	yawhg.ConfigYawhg(yawhg.Options{
		Enabled:    false,
		AppVersion: "test",
		LogLevel:   strings.Title(yawhg.InfoLevel.String()) + "Level",
	})

	// Pipe creates a synchronous in-memory pipe. It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
	r, w := io.Pipe()
	actualResult := new(bytes.Buffer)
	previousDestination := yawhg.Destination
	yawhg.Destination = w // send log output to the pipe instead of os.Stdout

	defer func() {
		yawhg.Destination = previousDestination
	}()

	go func() { // each Write to the PipeWriter blocks until one or more Reads from the PipeReader fully consume the written data

		cLogger := yawhg.NewLogger()

		cLogger.Infow(yawhg.Fields{
			"myField": "foo",
		})

		cLogger.Infow(yawhg.Fields{
			"additionalField": "bar",
		})

		w.Close()
	}()

	actualResult.ReadFrom(r) // blocks until data is put into the pipe

	for _, result := range want {
		if !strings.Contains(actualResult.String(), string(result)) {
			t.Fatalf("expected %v to contain %v", actualResult, result)
		}
	}
}

func TestLogConcurrently(t *testing.T) {
	yawhg.ConfigYawhg(yawhg.Options{
		Enabled:    false,
		AppVersion: "test",
		LogLevel:   "InfoLevel",
	})

	payload := yawhg.Fields{
		"Test": "Foo",
	}

	numTests := 100

	stringChan := make(chan string, numTests)

	t.Run("concurrency test", func(t *testing.T) {
		// test 100 concurrent logs to see if we get a panic
		wg := sync.WaitGroup{}
		for i := 0; i < numTests; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				msg := fmt.Sprintf("log number %d", i)
				yawhg.WithFields(payload).Info(msg)
				stringChan <- msg
			}(i)

		}
		wg.Wait()

		msgMap := make(map[string]struct{})

		for i := 0; i < numTests; i++ {
			msg := <-stringChan
			msgMap[msg] = struct{}{}
		}

		mapLen := len(msgMap)

		// confirm that no logs have collided and overwritten each other
		if mapLen != numTests {
			t.Fatalf("expected %d logs, but got %d", numTests, mapLen)
		}
	})
}

func BenchmarkYawhg(b *testing.B) {
	yawhg.ConfigYawhg(yawhg.Options{
		Enabled:    false,
		AppVersion: "test",
		LogLevel:   "InfoLevel",
	})

	payload := yawhg.Fields{
		"Test": "Foo",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		yawhg.WithFields(payload).Info("benchmark log")
	}
}
