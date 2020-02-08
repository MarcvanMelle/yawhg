## Yawhg is a structured logger.

## Instructions
Configure the logger at the earliest entrypoint in your app for which logs are necessary, e.g. your config package or main.go.
It is recommended that these initialization settings be passed in as environment variables, so that the logger can be disabled in the test
environment, for example.
```
yawhg.ConfigYawhg(yawhg.Options{
	Enabled:    true,  // optional, defaults to true
	AppVersion: "20180525", // or whatever versioning scheme you prefer
	LogLevel: "InfoLevel",
})
```

Please see the example folder for examples of how to use the logger and the output of those examples.
Capabilities include a logrus-style logger, a shorthand multi-field logger, a simple string logger, and a cumulative logger.

## Log Levels
Setting the log level means that only logs of that severity level and higher will be output.  The ascending order of
levels are as follows:

DebugLevel
InfoLevel
ErrorLevel

In other words, if you set a log level of "InfoLevel," then debug logs will not be output, but info and error logs will be.

If you do not specify the log level, a default of "InfoLevel" will be used.

## Benchmark
Run the benchmark with
```
go test ./... -run=SKIPTESTS -bench=.
```
Current results:
```
BenchmarkYawhg-8          2000000               823 ns/op             455 B/op          8 allocs/op
```

## TO DO
1. Increase efficiency of the logger


