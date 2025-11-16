# go-test-logger

Test logging utilities for Ginkgo/Gomega BDD tests - capture and validate test logs while suppressing expected error logs from test output.

## Overview

`go-test-logger` provides utilities for handling test logs in Ginkgo/Gomega test suites. It solves the common problem of expected error logs cluttering test output while ensuring unexpected errors remain visible for debugging.

## Features

- **ExpectErrorLog**: Capture and validate expected error patterns, hide matching logs
- **ConfigureTestLogging**: Suite-level logging configuration with sensible defaults
- **WithCapturedLogger**: Manual log capture for custom validation
- **AssertNoErrorLogs**: Negative assertions for successful operations
- **Pattern matching**: Validate log patterns while hiding expected output
- **JSON support**: Full support for structured JSON log validation
- **Gomega integration**: Works seamlessly with Gomega matchers
- **Context-aware**: Respects LOG_LEVEL environment variable for debugging

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
    "github.com/JohnPlummer/go-test-logger"
)

func TestMyPackage(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "My Package Suite")
}

var _ = BeforeSuite(func() {
    // Configure test logging - suppresses INFO/WARN, shows ERROR
    testlogger.ConfigureTestLogging()
})

var _ = Describe("API Client", func() {
    It("should handle rate limit errors gracefully", func() {
        // Expected errors are hidden from output, unexpected errors are shown
        testlogger.ExpectErrorLog(func(logger *slog.Logger) {
            client := NewClient(logger)
            err := client.CallAPI() // Will log "rate limit exceeded"
            Expect(err).To(HaveOccurred())
        }, "rate limit exceeded", "status=429")
    })

    It("should complete successfully without errors", func() {
        logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)
        client := NewClient(logger)

        err := client.ProcessData()
        Expect(err).NotTo(HaveOccurred())

        // Verify no ERROR logs were produced
        testlogger.AssertNoErrorLogs(buffer)
    })
})
```

## Core Concepts

### The Problem

When testing error handling code, expected error logs clutter test output making it hard to spot actual problems:

```
# Without go-test-logger
time=2025-11-16T11:00:00.000Z level=ERROR msg="rate limit exceeded" status=429
time=2025-11-16T11:00:01.000Z level=ERROR msg="connection timeout" retry=1
time=2025-11-16T11:00:02.000Z level=ERROR msg="invalid token" auth=failed
... (100 more expected errors in your test suite)
```

This noise makes it impossible to distinguish between:

- **Expected errors** (part of normal test flow)
- **Unexpected errors** (actual bugs that need attention)

### The Solution

`go-test-logger` filters expected errors while showing unexpected ones:

```go
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    // Test code that produces expected error logs
    client.CallRateLimitedAPI()
}, "rate limit exceeded") // Pattern to hide from output
```

**Result:**

- ✅ Expected "rate limit exceeded" logs: **HIDDEN** (validated silently)
- ❌ Unexpected logs (bugs): **SHOWN** in stderr for debugging

### Suite-Level Configuration

Configure logging once in `BeforeSuite`:

```go
var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
    // By default: suppresses INFO/WARN, shows ERROR
    // Set LOG_LEVEL=DEBUG for verbose output during debugging
})
```

This provides clean test output by default, with easy debugging when needed.

## API Reference

### ExpectErrorLog

Runs a test function with a captured logger and validates expected error patterns.

**Signature:**

```go
func ExpectErrorLog(testFunc func(*slog.Logger), expectedPatterns ...string)
```

**Behavior:**

- Expected logs (matching patterns): **HIDDEN** from output
- Unexpected logs (not matching): **SHOWN** to stderr
- Test fails if expected patterns not found (Gomega assertion)

**Example:**

```go
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    service := NewService(logger)
    err := service.ProcessInvalidData()
    Expect(err).To(HaveOccurred())
}, "validation failed", "invalid email format")
```

**When to use:**

- Testing error handling code paths
- Validating specific error log patterns
- Hiding expected errors from test output

### ExpectErrorLogJSON

Like `ExpectErrorLog` but uses JSON output format for validating structured log fields.

**Signature:**

```go
func ExpectErrorLogJSON(testFunc func(*slog.Logger), expectedPatterns ...string)
```

**Example:**

```go
testlogger.ExpectErrorLogJSON(func(logger *slog.Logger) {
    service := NewService(logger)
    service.ProcessData()
}, `"level":"ERROR"`, `"msg":"processing failed"`, `"user_id":"123"`)
```

**When to use:**

- Validating structured log fields
- Testing JSON log output
- Verifying specific field values in logs

### WithCapturedLogger

Creates a logger that writes to a buffer for manual validation.

**Signature:**

```go
func WithCapturedLogger(level slog.Level) (*slog.Logger, *gbytes.Buffer)
```

**Example:**

```go
logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

service := NewService(logger)
service.ProcessData()

// Manual validation with Gomega matchers
Expect(buffer).To(gbytes.Say("processing started"))
Expect(buffer).To(gbytes.Say("items processed"))
Expect(buffer).To(gbytes.Say("count=42"))
```

**When to use:**

- Custom log validation logic
- Complex assertions beyond pattern matching
- Fine-grained control over log validation

### WithCapturedJSONLogger

Creates a JSON logger for validating structured log output.

**Signature:**

```go
func WithCapturedJSONLogger(level slog.Level) (*slog.Logger, *gbytes.Buffer)
```

**Example:**

```go
logger, buffer := testlogger.WithCapturedJSONLogger(slog.LevelInfo)

handler := NewHandler(logger)
handler.HandleRequest(req)

Expect(buffer).To(gbytes.Say(`"request_id":"abc123"`))
Expect(buffer).To(gbytes.Say(`"method":"POST"`))
Expect(buffer).To(gbytes.Say(`"status":200`))
```

**When to use:**

- Validating JSON log structure
- Testing structured logging
- Verifying JSON field values

### AssertNoErrorLogs

Validates that no ERROR level logs were produced.

**Signature:**

```go
func AssertNoErrorLogs(buffer *gbytes.Buffer)
```

**Example:**

```go
logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)
service := NewService(logger)

err := service.ProcessValidData()
Expect(err).NotTo(HaveOccurred())

// Verify no ERROR logs (for both text and JSON)
testlogger.AssertNoErrorLogs(buffer)
```

**When to use:**

- Testing successful code paths
- Ensuring error-free execution
- Negative assertions (proving absence of errors)

### ConfigureTestLogging

Sets up slog for test suites with sensible defaults.

**Signature:**

```go
func ConfigureTestLogging()
```

**Default behavior:**

- Suppresses INFO and WARN messages
- Shows ERROR logs to stderr for debugging
- Respects LOG_LEVEL environment variable

**LOG_LEVEL values:**

- `DEBUG`: Shows all logs (most verbose)
- `INFO`: Shows INFO and above
- `WARN`: Shows WARN and above
- `ERROR`: Shows ERROR only
- (default): Suppresses INFO/WARN, shows ERROR

**Example:**

```go
var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
    // Additional suite setup...
})
```

**Shell usage:**

```bash
# Default: quiet output (ERROR only)
ginkgo run ./...

# Verbose: see all logs for debugging
LOG_LEVEL=DEBUG ginkgo run ./pkg/client

# Moderate: see INFO and above
LOG_LEVEL=INFO ginkgo run ./...
```

## Usage Patterns

### Pattern 1: Testing Error Handling

When testing code that logs expected errors:

```go
It("should handle database connection errors", func() {
    testlogger.ExpectErrorLog(func(logger *slog.Logger) {
        repo := NewRepository(logger, invalidConfig)
        err := repo.Connect()
        Expect(err).To(HaveOccurred())
    }, "connection failed", "connection refused")
})
```

### Pattern 2: Testing Success Paths

When verifying code completes without errors:

```go
It("should process valid data successfully", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)
    processor := NewProcessor(logger)

    err := processor.Process(validData)
    Expect(err).NotTo(HaveOccurred())
    testlogger.AssertNoErrorLogs(buffer)
})
```

### Pattern 3: Testing Structured Logs

When validating JSON log fields:

```go
It("should log user actions with structured data", func() {
    testlogger.ExpectErrorLogJSON(func(logger *slog.Logger) {
        tracker := NewTracker(logger)
        tracker.RecordAction("login", "user123")
    }, `"action":"login"`, `"user_id":"user123"`, `"timestamp"`)
})
```

### Pattern 4: Custom Validation

When you need fine-grained control:

```go
It("should log processing steps in correct order", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelInfo)
    pipeline := NewPipeline(logger)

    pipeline.Execute(data)

    // Validate specific log sequence
    Expect(buffer).To(gbytes.Say("stage 1: validation"))
    Expect(buffer).To(gbytes.Say("stage 2: transformation"))
    Expect(buffer).To(gbytes.Say("stage 3: persistence"))
})
```

## Best Practices

### 1. Configure Logging in BeforeSuite

Always configure test logging once at the suite level:

```go
var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
})
```

**Why:**

- Consistent logging behavior across all tests
- Clean test output by default
- Easy debugging with LOG_LEVEL environment variable

### 2. Use Specific Patterns

Provide specific patterns to validate exact error messages:

```go
// ✅ Good: Specific patterns
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    client.Connect()
}, "connection failed", "timeout", "host=localhost")

// ❌ Bad: Vague patterns
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    client.Connect()
}, "error")
```

### 3. Test Both Success and Failure Paths

Validate both error conditions and error-free execution:

```go
It("should handle invalid input", func() {
    testlogger.ExpectErrorLog(func(logger *slog.Logger) {
        // Test error path
    }, "validation failed")
})

It("should process valid input successfully", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)
    // Test success path
    testlogger.AssertNoErrorLogs(buffer)
})
```

### 4. Match Log Format to Code

Use JSON validation for code that uses JSON handlers:

```go
// If your code uses slog.NewJSONHandler
testlogger.ExpectErrorLogJSON(func(logger *slog.Logger) {
    service.Process()
}, `"level":"ERROR"`, `"msg":"failed"`)

// If your code uses slog.NewTextHandler
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    service.Process()
}, "level=ERROR", "msg=failed")
```

### 5. Escape Special Regex Characters

When patterns contain regex special characters, escape them:

```go
testlogger.ExpectErrorLog(func(logger *slog.Logger) {
    parser.Parse("[invalid]")
}, "\\[invalid\\]") // Escape brackets
```

## Migration Guide

### From Raw slog Testing

**Before** (logs clutter output):

```go
It("should handle errors", func() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    client := NewClient(logger)
    err := client.CallAPI()
    Expect(err).To(HaveOccurred())
    // ERROR logs appear in test output ❌
})
```

**After** (clean output):

```go
var _ = BeforeSuite(func() {
    testlogger.ConfigureTestLogging()
})

It("should handle errors", func() {
    testlogger.ExpectErrorLog(func(logger *slog.Logger) {
        client := NewClient(logger)
        err := client.CallAPI()
        Expect(err).To(HaveOccurred())
    }, "rate limit exceeded")
    // Expected ERROR logs hidden ✅
})
```

### From Manual Buffer Capture

**Before** (verbose setup):

```go
It("should log operations", func() {
    var buf bytes.Buffer
    logger := slog.New(slog.NewTextHandler(&buf, nil))
    service := NewService(logger)
    service.Process()

    output := buf.String()
    Expect(output).To(ContainSubstring("processing"))
})
```

**After** (concise):

```go
It("should log operations", func() {
    logger, buffer := testlogger.WithCapturedLogger(slog.LevelInfo)
    service := NewService(logger)
    service.Process()

    Expect(buffer).To(gbytes.Say("processing"))
})
```

## Testing

Run tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-coverage
```

Run tests with race detection:

```bash
make test-race
```

Check code formatting and linting:

```bash
make check
```

## Contributing

Contributions welcome! Please ensure:

1. Tests pass: `make test`
2. Coverage stays >90%: `make test-coverage`
3. Linting passes: `make lint`
4. Code is formatted: `make fmt`
5. Documentation is updated

## License

MIT License - see [LICENSE](./LICENSE) file for details.

## Related Packages

- [Ginkgo](https://github.com/onsi/ginkgo): BDD testing framework
- [Gomega](https://github.com/onsi/gomega): Matcher library
- [gbytes](https://github.com/onsi/gomega/tree/master/gbytes): Buffer testing utilities
- [slog](https://pkg.go.dev/log/slog): Structured logging (Go standard library)
