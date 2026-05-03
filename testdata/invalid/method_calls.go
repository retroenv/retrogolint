package testdata

import "net"

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct{}

func (l *Logger) Info(msg string, fields ...Field) {}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Stringer(key string, value interface{ String() string }) Field {
	return Field{Key: key, Value: value.String()}
}

func ExampleMethodCalls() {
	logger := &Logger{}
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}

	// Bad: using .String() method with log.String (should use log.Stringer)
	logger.Info("Connection", String("address", addr.String()))
}
