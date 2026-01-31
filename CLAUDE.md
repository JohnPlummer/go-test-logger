# go-test-logger

Test logging utilities for Ginkgo/Gomega BDD tests using `log/slog`.

**Module**: `github.com/JohnPlummer/go-test-logger`

## API

| Function | Purpose |
|----------|---------|
| `ConfigureTestLogging()` | Suite-level log filtering (call in BeforeSuite) |
| `ExpectErrorLog(fn, patterns...)` | Capture slog output, validate patterns, hide matched logs |
| `ExpectErrorLogJSON(fn, patterns...)` | Same as above with JSON format |
| `WithCapturedLogger(level)` | Returns `(*slog.Logger, *gbytes.Buffer)` for manual validation |
| `WithCapturedJSONLogger(level)` | JSON variant of above |
| `AssertNoErrorLogs(buffer)` | Assert no ERROR level logs in buffer |

## Usage

### Suite Setup

```go
BeforeSuite(func() { testlogger.ConfigureTestLogging() })
```

### Validate Expected Errors

```go
testlogger.ExpectErrorLog(func(*slog.Logger) {
    service := NewService()
    err := service.Process()
    Expect(err).To(HaveOccurred())
}, "processing failed", "invalid input")
```

### Manual Buffer Access

Only use when you need direct access to the buffer for custom assertions or log sequence validation.

```go
logger, buffer := testlogger.WithCapturedLogger(slog.LevelInfo)
service := NewService(logger)  // must pass logger
service.Process()
Expect(buffer).To(gbytes.Say("started"))
```

## LOG_LEVEL

| Value | Effect |
|-------|--------|
| DEBUG | All logs to stderr |
| INFO | INFO and above |
| WARN | WARN and above |
| ERROR | ERROR only |
| (unset) | Suppress INFO/WARN, show ERROR |

## Commands

```bash
make test           # Run tests
make check          # Format, lint, test
make coverage       # Generate coverage report
make lint           # Run linter
LOG_LEVEL=DEBUG make test  # Debug logging
```

## Notes

- `ExpectErrorLog` captures all global slog calls automatically
- `WithCapturedLogger` requires passing the logger to code under test
- Patterns use regex (escape special chars: `\\[`, `\\(`)
