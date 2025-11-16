package testlogger

import (
	"log/slog"
	"os"
)

// ConfigureTestLogging sets up slog for test suites with sensible defaults
// for Ginkgo/Gomega BDD testing.
//
// By default, suppresses INFO and WARN messages but shows ERROR for debugging.
//
// The LOG_LEVEL environment variable controls logging verbosity:
//   - DEBUG: Shows all logs to stderr (most verbose)
//   - INFO: Shows INFO and above to stderr
//   - WARN: Shows WARN and above to stderr
//   - ERROR: Shows ERROR only to stderr
//   - (default): Shows ERROR only, suppresses INFO and WARN
//
// This should be called in BeforeSuite to configure logging for the entire test suite:
//
//	var _ = BeforeSuite(func() {
//	    testlogger.ConfigureTestLogging()
//	    // Suite setup continues...
//	})
func ConfigureTestLogging() {
	// By default, suppress INFO and WARN but allow ERROR for debugging
	// slog levels: DEBUG=-4, INFO=0, WARN=4, ERROR=8
	// Setting to 7 (just below ERROR) suppresses INFO and WARN
	logLevel := slog.Level(7) // Just below ERROR to suppress INFO and WARN
	output := os.Stderr       // Show ERROR logs to stderr for debugging

	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		switch lvl {
		case "DEBUG":
			logLevel = slog.LevelDebug
			output = os.Stderr // Show output for debugging
		case "INFO":
			logLevel = slog.LevelInfo
			output = os.Stderr
		case "WARN":
			logLevel = slog.LevelWarn
			output = os.Stderr
		case "ERROR":
			// Show ERROR logs only
			logLevel = slog.LevelError
			output = os.Stderr
		}
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(output, opts))
	slog.SetDefault(logger)
}
