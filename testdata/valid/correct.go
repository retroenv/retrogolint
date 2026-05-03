package testdata

type Field struct {
	Key   string
	Value interface{}
}

type GoodLogger struct{}

func (l *GoodLogger) Info(msg string, fields ...Field) {}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func ExampleCorrect() {
	logger := &GoodLogger{}

	// All correct: uppercase messages and snake_case fields
	logger.Info("Starting server")
	logger.Info("Loading file", String("file_name", "test.txt"))
	logger.Info("Process complete", String("process_id", "123"))
}
