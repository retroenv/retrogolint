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

func StringFunc(key string, fn func() string) Field {
	return Field{Key: key, Value: fn()}
}

func ExampleEfficiency() {
	logger := &Logger{}

	// Bad: calling expensive function directly (should use StringFunc for lazy evaluation)
	logger.Info("Process", String("result", expensiveOperation()))
}

func expensiveOperation() string {
	return "result"
}
