# go-test-logger

Test logging utilities for Ginkgo/Gomega BDD tests using `log/slog`.

**Module**: `github.com/JohnPlummer/go-test-logger`

## Core API

| Function | Purpose |
|----------|---------|
| `ConfigureTestLogging()` | Suite-level global slog config (call in BeforeSuite) |
| `ExpectErrorLog(fn, patterns...)` | Capture logs, validate patterns, hide matched output |
| `ExpectErrorLogJSON(fn, patterns...)` | Same as above with JSON format |
| `WithCapturedLogger(level)` | Returns `(*slog.Logger, *gbytes.Buffer)` for manual validation |
| `WithCapturedJSONLogger(level)` | JSON variant of above |
| `AssertNoErrorLogs(buffer)` | Assert no ERROR level logs in buffer |

## Two Approaches

1. **Logger Injection** (preferred): Pass captured logger to code under test

   ```go
   ExpectErrorLog(func(logger *slog.Logger) {
       client := NewClient(logger)
       err := client.CallAPI()
       Expect(err).To(HaveOccurred())
   }, "rate limit exceeded", "status=429")
   ```

2. **Global Logger**: Configure default slog in BeforeSuite

   ```go
   BeforeSuite(func() { testlogger.ConfigureTestLogging() })
   ```

## LOG_LEVEL Environment Variable

| Value | Effect |
|-------|--------|
| DEBUG | All logs to stderr |
| INFO | INFO and above |
| WARN | WARN and above |
| ERROR | ERROR only |
| (unset) | Suppress INFO/WARN, show ERROR |

## Commands

```bash
go test ./...              # Run tests
go test -cover ./...       # With coverage
go test -v ./...           # Verbose
LOG_LEVEL=DEBUG go test    # Debug logging
```

## Shared Package Rules

- All changes require tests
- No project-specific code
- Patterns use regex (escape special chars: `\\[`, `\\(`)
