package testdata

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct{}

func (l *Logger) Info(msg string, fields ...Field) {}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func ExampleFieldCasing() {
	logger := &Logger{}

	// Bad: camelCase field names (should be "file_name")
	logger.Info("Loading file", String("fileName", "test.txt"))

	// Bad: PascalCase field names (should be "process_id")
	logger.Info("Process", Int("ProcessId", 123))

	// Bad: camelCase field names (should be "user_name")
	logger.Info("User", String("userName", "alice"))
}
