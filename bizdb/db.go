package bizdb

// Logger interface for logging SQL operations
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
}
