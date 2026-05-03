package testdata

import "fmt"

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct{}

func (l *Logger) Info(msg string, fields ...Field) {}
func (l *Logger) Debug(msg string, fields ...Field) {}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Hex(key string, value int) Field {
	return Field{Key: key, Value: fmt.Sprintf("%x", value)}
}

func ExampleHexFormatting() {
	logger := &Logger{}
	address := 0x1234
	value := 0xABCD

	// Bad: using fmt.Sprintf with %x (should use log.Hex)
	logger.Info("Memory", String("address", fmt.Sprintf("0x%04X", address)))

	// Bad: using fmt.Sprintf with %x (should use log.Hex)
	logger.Debug("Value", String("data", fmt.Sprintf("%x", value)))
}
