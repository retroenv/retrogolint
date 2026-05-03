package testdata

type Logger struct{}

func (l *Logger) Info(msg string)  {}
func (l *Logger) Debug(msg string) {}
func (l *Logger) Error(msg string) {}

func ExampleCapitalization() {
	logger := &Logger{}

	// Bad: lowercase message (should be "Starting server")
	logger.Info("starting server")

	// Bad: lowercase message (should be "Loading configuration")
	logger.Debug("loading configuration")

	// Bad: lowercase message (should be "Connection failed")
	logger.Error("connection failed")
}
