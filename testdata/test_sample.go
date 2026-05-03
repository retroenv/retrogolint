package testdata

type TestLogger struct{}

func (l *TestLogger) Info(msg string) {}

func TestSample() {
	logger := &TestLogger{}
	logger.Info("test message")
}
