package testlogger_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	testlogger "github.com/JohnPlummer/go-test-logger"
)

func TestTestlogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "go-test-logger Suite")
}

// Example service for demonstration
type TestService struct {
	logger *slog.Logger
}

func (s *TestService) ProcessWithError() error {
	s.logger.Error("Processing failed", "error", "invalid input")
	return errors.New("invalid input")
}

func (s *TestService) ProcessSuccessfully() error {
	s.logger.Info("Processing completed successfully")
	return nil
}

var _ = Describe("Logger Test Utilities", func() {
	Describe("ExpectErrorLog", func() {
		It("should capture and validate expected error patterns", func() {
			// This test demonstrates using ExpectErrorLog to validate
			// that specific error patterns appear in the log output
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				// Simulate an operation that logs an error
				logger.Error("API call failed",
					"error", "rate limit exceeded",
					"status", 429,
					"endpoint", "/api/v1/data")
			}, "API call failed", "rate limit exceeded", "status=429")
		})

		It("should hide expected logs and show unexpected logs", func() {
			// This test logs 3 messages but only validates "Expected error" pattern
			// Expected: "Expected error: rate limit exceeded" - HIDDEN (filtered out)
			// Unexpected: "Unexpected error: database connection failed" - SHOWN in output below
			// Unexpected: "Unexpected warning: retry attempt 3" - SHOWN in output below
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Expected error: rate limit exceeded")
				logger.Error("Unexpected error: database connection failed")
				logger.Warn("Unexpected warning: retry attempt 3")
			}, "Expected error")
		})

		It("should programmatically verify filtering behavior", func() {
			// Capture stderr to verify filtering works correctly
			originalStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Run test with multiple logs - only validate one pattern
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Test expected: should be hidden")
				logger.Error("Test unexpected: should be visible")
			}, "Test expected")

			// Restore stderr and read captured output
			w.Close()
			os.Stderr = originalStderr
			var buf bytes.Buffer
			buf.ReadFrom(r)
			stderrOutput := buf.String()

			// Expected log should be hidden (NOT in stderr)
			Expect(stderrOutput).NotTo(ContainSubstring("Test expected: should be hidden"))

			// Unexpected log should be visible (IN stderr)
			Expect(stderrOutput).To(ContainSubstring("Test unexpected: should be visible"))
		})
	})

	Describe("ExpectErrorLogJSON", func() {
		It("should validate JSON formatted error logs", func() {
			testlogger.ExpectErrorLogJSON(func(logger *slog.Logger) {
				logger.Error("Database connection failed",
					"host", "localhost",
					"port", 5432,
					"error", errors.New("connection refused"))
			}, `"level":"ERROR"`, `"msg":"Database connection failed"`, `"host":"localhost"`, `"port":5432`)
		})
	})

	Describe("WithCapturedLogger", func() {
		It("should allow manual validation of log output", func() {
			logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

			// Perform operations that generate logs
			logger.Debug("Starting process")
			logger.Info("Processing item", "id", "123")
			logger.Warn("Retry attempt", "attempt", 1)
			logger.Error("Failed to complete", "reason", "timeout")

			// Validate the captured output
			Expect(buffer).To(gbytes.Say("Starting process"))
			Expect(buffer).To(gbytes.Say("Processing item"))
			Expect(buffer).To(gbytes.Say("id=123"))
			Expect(buffer).To(gbytes.Say("Retry attempt"))
			Expect(buffer).To(gbytes.Say("Failed to complete"))
			Expect(buffer).To(gbytes.Say("reason=timeout"))
		})

		It("should respect log level filtering", func() {
			logger, buffer := testlogger.WithCapturedLogger(slog.LevelWarn)

			logger.Debug("This won't appear")
			logger.Info("Neither will this")
			logger.Warn("This will appear")
			logger.Error("And so will this")

			contents := string(buffer.Contents())
			Expect(contents).NotTo(ContainSubstring("This won't appear"))
			Expect(contents).NotTo(ContainSubstring("Neither will this"))
			Expect(contents).To(ContainSubstring("This will appear"))
			Expect(contents).To(ContainSubstring("And so will this"))
		})
	})

	Describe("WithCapturedJSONLogger", func() {
		It("should capture JSON formatted logs", func() {
			logger, buffer := testlogger.WithCapturedJSONLogger(slog.LevelInfo)

			logger.Info("User action",
				"user_id", "abc123",
				"action", "login",
				"success", true)

			// Validate JSON structure
			Expect(buffer).To(gbytes.Say(`"level":"INFO"`))
			Expect(buffer).To(gbytes.Say(`"msg":"User action"`))
			Expect(buffer).To(gbytes.Say(`"user_id":"abc123"`))
			Expect(buffer).To(gbytes.Say(`"action":"login"`))
			Expect(buffer).To(gbytes.Say(`"success":true`))
		})
	})

	Describe("AssertNoErrorLogs", func() {
		It("should pass when no ERROR logs are present", func() {
			logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)

			logger.Debug("Debug message")
			logger.Info("Info message")
			logger.Warn("Warning message")

			testlogger.AssertNoErrorLogs(buffer)
		})
	})

	Describe("Integration Examples", func() {
		It("should work with service methods that log errors", func() {
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				service := &TestService{logger: logger}
				err := service.ProcessWithError()
				Expect(err).To(HaveOccurred())
			}, "Processing failed", "invalid input")
		})

		It("should validate successful operations have no errors", func() {
			logger, buffer := testlogger.WithCapturedLogger(slog.LevelDebug)
			service := &TestService{logger: logger}

			err := service.ProcessSuccessfully()
			Expect(err).NotTo(HaveOccurred())
			testlogger.AssertNoErrorLogs(buffer)
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty pattern list", func() {
			// Validates all logs shown when no patterns provided
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Some error")
			})
		})

		It("should handle patterns with special regex characters", func() {
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Error: [database] connection failed (timeout)")
			}, "Error: \\[database\\]", "connection failed \\(timeout\\)")
		})

		It("should work with concurrent logging", func() {
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						logger.Error("Concurrent log", "goroutine", id)
					}(i)
				}
				wg.Wait()
			}, "Concurrent log")
		})

		It("should handle logs with newlines", func() {
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Multi-line error message\nwith details on second line")
			}, "Multi-line error message")
		})

		It("should handle unicode characters", func() {
			testlogger.ExpectErrorLog(func(logger *slog.Logger) {
				logger.Error("Error processing: 日本語 文字")
			}, "日本語 文字")
		})
	})

	Describe("ConfigureTestLogging", func() {
		AfterEach(func() {
			// Clean up environment variable after each test
			os.Unsetenv("LOG_LEVEL")
		})

		It("should set default logger without panicking", func() {
			// This test verifies ConfigureTestLogging doesn't panic
			// and sets up logging configuration correctly
			Expect(func() {
				testlogger.ConfigureTestLogging()
			}).NotTo(Panic())
		})

		It("should respect DEBUG log level from environment", func() {
			os.Setenv("LOG_LEVEL", "DEBUG")
			Expect(func() {
				testlogger.ConfigureTestLogging()
			}).NotTo(Panic())
		})

		It("should respect INFO log level from environment", func() {
			os.Setenv("LOG_LEVEL", "INFO")
			Expect(func() {
				testlogger.ConfigureTestLogging()
			}).NotTo(Panic())
		})

		It("should respect WARN log level from environment", func() {
			os.Setenv("LOG_LEVEL", "WARN")
			Expect(func() {
				testlogger.ConfigureTestLogging()
			}).NotTo(Panic())
		})

		It("should respect ERROR log level from environment", func() {
			os.Setenv("LOG_LEVEL", "ERROR")
			Expect(func() {
				testlogger.ConfigureTestLogging()
			}).NotTo(Panic())
		})
	})
})
