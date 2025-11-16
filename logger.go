// Package testlogger provides utilities for capturing and validating test logs
// while suppressing expected error logs from test output.
//
// The package integrates with Ginkgo/Gomega BDD testing and log/slog to provide:
//   - ExpectErrorLog: Capture and validate expected error patterns
//   - ConfigureTestLogging: Suite-level logging configuration
//   - WithCapturedLogger: Manual log capture for custom validation
//   - AssertNoErrorLogs: Negative assertions for successful operations
//
// Example usage:
//
//	var _ = BeforeSuite(func() {
//	    testlogger.ConfigureTestLogging()
//	})
//
//	It("should handle API errors gracefully", func() {
//	    testlogger.ExpectErrorLog(func(logger *slog.Logger) {
//	        client := NewClient(logger)
//	        err := client.CallAPI()
//	        Expect(err).To(HaveOccurred())
//	    }, "rate limit exceeded", "status=429")
//	})
package testlogger

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

// getLogLevel reads the LOG_LEVEL environment variable and returns the appropriate slog.Level.
// Defaults to slog.Level(7) which is just below ERROR to suppress INFO and WARN.
func getLogLevel() slog.Level {
	if level, ok := map[string]slog.Level{
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}[os.Getenv("LOG_LEVEL")]; ok {
		return level
	}
	return slog.Level(7) // Just below ERROR to suppress INFO and WARN
}

// expectErrorLogWithHandler is a helper that consolidates the common logic
// for capturing and validating error logs with different handler types.
func expectErrorLogWithHandler(
	handlerFactory func(io.Writer, *slog.HandlerOptions) slog.Handler,
	testFunc func(*slog.Logger),
	expectedPatterns ...string,
) {
	buffer := gbytes.NewBuffer()
	var capturedOutput bytes.Buffer
	writer := io.MultiWriter(buffer, &capturedOutput)
	logger := slog.New(handlerFactory(writer, &slog.HandlerOptions{
		Level: getLogLevel(),
	}))

	// Run test function with captured logger
	testFunc(logger)

	// Validate expected patterns appear in the log output
	for _, pattern := range expectedPatterns {
		Expect(buffer).To(gbytes.Say(pattern),
			"Expected error log pattern not found: %s", pattern)
	}

	// Display only unexpected logs (lines not matching any expected pattern)
	lines := strings.Split(capturedOutput.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		isExpected := false
		for _, pattern := range expectedPatterns {
			if strings.Contains(line, pattern) {
				isExpected = true
				break
			}
		}
		if !isExpected {
			fmt.Fprintln(os.Stderr, line)
		}
	}
}

// ExpectErrorLog runs a test function with a captured logger and validates
// that expected error patterns appear in the log output.
//
// This function integrates with Gomega's assertion framework to provide
// clear test failures when expected log patterns are not found.
//
// Expected logs (matching validation patterns) are hidden from output.
// Unexpected logs are displayed to stderr for debugging.
//
// Usage:
//
//	ExpectErrorLog(func(logger *slog.Logger) {
//	    client := NewClient(logger)
//	    err := client.CallAPI() // This will trigger expected error
//	    Expect(err).To(HaveOccurred())
//	}, "rate limit exceeded", "status=429")
func ExpectErrorLog(testFunc func(*slog.Logger), expectedPatterns ...string) {
	expectErrorLogWithHandler(
		func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewTextHandler(w, opts)
		},
		testFunc,
		expectedPatterns...,
	)
}

// ExpectErrorLogJSON is like ExpectErrorLog but uses JSON output format,
// which is useful for validating structured log fields.
//
// This function integrates with Gomega's assertion framework to provide
// clear test failures when expected log patterns are not found.
//
// Expected logs (matching validation patterns) are hidden from output.
// Unexpected logs are displayed to stderr for debugging.
//
// Usage:
//
//	ExpectErrorLogJSON(func(logger *slog.Logger) {
//	    service := NewService(logger)
//	    service.ProcessInvalidData()
//	}, `"level":"ERROR"`, `"msg":"validation failed"`, `"field":"email"`)
func ExpectErrorLogJSON(testFunc func(*slog.Logger), expectedPatterns ...string) {
	expectErrorLogWithHandler(
		func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		},
		testFunc,
		expectedPatterns...,
	)
}

// WithCapturedLogger creates a logger that writes to a gbytes.Buffer,
// allowing manual validation of log output using Gomega matchers.
//
// This provides fine-grained control over log validation when the automatic
// pattern matching in ExpectErrorLog is not sufficient.
//
// Usage:
//
//	logger, buffer := WithCapturedLogger(slog.LevelError)
//	service := NewService(logger)
//	service.ProcessData()
//	Expect(buffer).To(gbytes.Say("processing started"))
func WithCapturedLogger(level slog.Level) (*slog.Logger, *gbytes.Buffer) {
	buffer := gbytes.NewBuffer()
	logger := slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{
		Level: level,
	}))
	return logger, buffer
}

// WithCapturedJSONLogger creates a JSON logger that writes to a gbytes.Buffer,
// useful for validating structured log fields.
//
// This is particularly useful when testing code that relies on structured
// logging and you need to validate specific JSON fields in the log output.
//
// Usage:
//
//	logger, buffer := WithCapturedJSONLogger(slog.LevelInfo)
//	handler := NewHandler(logger)
//	handler.HandleRequest(req)
//	Expect(buffer).To(gbytes.Say(`"request_id":"123"`))
func WithCapturedJSONLogger(level slog.Level) (*slog.Logger, *gbytes.Buffer) {
	buffer := gbytes.NewBuffer()
	logger := slog.New(slog.NewJSONHandler(buffer, &slog.HandlerOptions{
		Level: level,
	}))
	return logger, buffer
}

// AssertNoErrorLogs validates that no ERROR level logs were produced.
// Useful for ensuring operations complete successfully without errors.
//
// This function checks for both text and JSON formatted error logs.
//
// Usage:
//
//	logger, buffer := WithCapturedJSONLogger(slog.LevelDebug)
//	service := NewService(logger)
//	service.ProcessValidData()
//	AssertNoErrorLogs(buffer)
func AssertNoErrorLogs(buffer *gbytes.Buffer) {
	contents := buffer.Contents()
	// Check for both text and JSON error indicators
	Expect(string(contents)).NotTo(ContainSubstring("level=ERROR"),
		"Unexpected ERROR log found in output")
	Expect(string(contents)).NotTo(ContainSubstring(`"level":"ERROR"`),
		"Unexpected ERROR log found in JSON output")
}
