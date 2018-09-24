package dnsutil

type Logger interface {
	Error(tag, msg string, args ...interface{})
	Info(tag, msg string, args ...interface{})
	Debug(tag, msg string, args ...interface{})
}

type PrefixedLogger struct {
	prefix string
	logger Logger
}

var _ Logger = PrefixedLogger{}

func NewPrefixedLogger(prefix string, logger Logger) PrefixedLogger {
	return PrefixedLogger{prefix, logger}
}

func (l PrefixedLogger) Error(tag, msg string, args ...interface{}) {
	l.logger.Error(tag, l.prefix+msg, args...)
}

func (l PrefixedLogger) Info(tag, msg string, args ...interface{}) {
	l.logger.Info(tag, l.prefix+msg, args...)
}

func (l PrefixedLogger) Debug(tag, msg string, args ...interface{}) {
	l.logger.Debug(tag, l.prefix+msg, args...)
}
