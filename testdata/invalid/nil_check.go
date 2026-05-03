package testdata

type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string) {}
func (l *SimpleLogger) Debug(msg string) {}

func ExampleNilCheck(logger *SimpleLogger) {
	// Bad: unnecessary nil check (logger methods are nil-safe)
	if logger != nil {
		logger.Info("Processing")
	}

	// Bad: another unnecessary nil check
	if logger != nil {
		logger.Debug("Status update")
	}

	// Bad: nil check in reverse order
	if nil != logger {
		logger.Info("Complete")
	}
}
