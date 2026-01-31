# go-test-logger

Test logging utilities for Ginkgo/Gomega BDD tests. Capture and validate log output while hiding expected errors from test output.

## Installation

```bash
go get github.com/JohnPlummer/go-test-logger
```

## Quick Start

```go
package mypackage_test

import (
    "log/slog"
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    testlogger "github.com/JohnPlummer/go-test-logger"
)

func TestMyPackage(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "My Package Suite")
}

var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
})

var _ = Describe("API Client", func() {
    It("handles rate limit errors", func() {
        testlogger.ExpectErrorLog(func(*slog.Logger) {
            client := NewClient()
            err := client.CallAPI()
            Expect(err).To(HaveOccurred())
        }, "rate limit exceeded", "status=429")
    })
})
```

## API

### ConfigureTestLogging

Call once in `BeforeSuite` to configure log levels for the test suite.

```go
var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
})
```

By default, suppresses INFO and WARN logs while showing ERROR. Control with `LOG_LEVEL` environment variable:

```bash
make test                   # Run tests (ERROR only by default)
LOG_LEVEL=DEBUG make test   # Show all logs
LOG_LEVEL=INFO make test    # Show INFO and above
```

### ExpectErrorLog

Captures all `slog` output during the test function. Validates that expected patterns appear, hides matching logs, shows unexpected logs to stderr.

```go
testlogger.ExpectErrorLog(func(*slog.Logger) {
    service := NewService()
    err := service.Process()
    Expect(err).To(HaveOccurred())
}, "processing failed", "invalid input")
```

The test fails if any expected pattern is not found in the log output.

### ExpectErrorLogJSON

Same as `ExpectErrorLog` but captures JSON-formatted logs.

```go
testlogger.ExpectErrorLogJSON(func(*slog.Logger) {
    service := NewService()
    service.Process()
}, `"level":"ERROR"`, `"msg":"failed"`, `"code":500`)
```

### WithCapturedLogger

Returns a logger and buffer for manual validation. Only use when you need direct access to the buffer for custom assertions or log sequence validation.

```go
logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

service := NewService(logger)
service.Process()

Expect(buffer).To(gbytes.Say("step 1"))
Expect(buffer).To(gbytes.Say("step 2"))
```

**Note:** Does not capture global `slog` calls. You must pass the returned logger to your code.

### WithCapturedJSONLogger

Same as `WithCapturedLogger` but returns a JSON-formatted logger.

```go
logger, buffer := testlogger.WithCapturedJSONLogger(slog.LevelInfo)

handler := NewHandler(logger)
handler.HandleRequest(req)

Expect(buffer).To(gbytes.Say(`"request_id":"123"`))
```

### AssertNoErrorLogs

Validates that no ERROR level logs were produced.

```go
logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

service := NewService(logger)
err := service.ProcessValidData()
Expect(err).NotTo(HaveOccurred())

testlogger.AssertNoErrorLogs(buffer)
```

## Examples

### Testing Error Handling

```go
It("logs connection errors", func() {
    testlogger.ExpectErrorLog(func(*slog.Logger) {
        db := NewDatabase(invalidConfig)
        err := db.Connect()
        Expect(err).To(HaveOccurred())
    }, "connection failed", "timeout")
})
```

### Validating Log Sequence

```go
It("logs processing steps in order", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelInfo)

    pipeline := NewPipeline(logger)
    pipeline.Execute(data)

    Expect(buffer).To(gbytes.Say("validation"))
    Expect(buffer).To(gbytes.Say("transformation"))
    Expect(buffer).To(gbytes.Say("persistence"))
})
```

### Testing Success Path

```go
It("completes without errors", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

    service := NewService(logger)
    err := service.Process(validData)

    Expect(err).NotTo(HaveOccurred())
    testlogger.AssertNoErrorLogs(buffer)
})
```

## Pattern Matching

Patterns are matched using `gbytes.Say()` which supports regular expressions. Escape special characters:

```go
testlogger.ExpectErrorLog(func(*slog.Logger) {
    parser.Parse("[invalid]")
}, "\\[invalid\\]")
```

## Changelog

### v1.2.0

`ExpectErrorLog` and `ExpectErrorLogJSON` now capture all `slog` output automatically. Code using `slog.Error()`, `slog.Info()`, etc. is captured without any setup.

## License

MIT
