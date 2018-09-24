package dstconn

type Logger interface {
	Error(tag, msg string, args ...interface{})
	Info(tag, msg string, args ...interface{})
	Debug(tag, msg string, args ...interface{})
}
