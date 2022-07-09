package logging

import (
	"fmt"
	"main/config"
	"main/global"
	"os"
	"path"
	"runtime"
	"time"

	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
)

// ContextKeyRequestID ...
var ContextKeyRequestID = "request_id"

// ContextKeyUserID ...
var ContextKeyUserID = "user_id"

// Logger is our logger instance abstraction.
type Logger struct {
	*logging.Logger
}

// Singleton StackDriver client.
var stackDriverClient *logging.Client

// StackDriverLogger is the singleton StackDriver logger instance.
var StackDriverLogger = &Logger{}

// Static configuration variables initalized at runtime.
var logLevel uint
var stackDriverEnabled bool
var gRPCConnectTimeout time.Duration
var projectID string

// Log levels.
const (
	logLevelFirst = iota
	logLevelCritical
	logLevelError
	logLevelWarn
	logLevelInfo
	logLevelDebug
	logLevelLast
)

// Log level to label string.
var logLabels = []string{
	"",
	"\x1b[0;37;41m  CRIT \x1b[m",
	"\x1b[0;30;41m ERROR \x1b[m",
	"\x1b[0;30;43m  WARN \x1b[m",
	"\x1b[0;30;47m  INFO \x1b[m",
	"\x1b[0;30;42m DEBUG \x1b[m",
	"",
}

// Log level to StackDriver severity.
var logSeverities = []logging.Severity{
	logging.Default,
	logging.Critical,
	logging.Error,
	logging.Warning,
	logging.Info,
	logging.Debug,
	logging.Default,
}

// init loads the logging configurations.
func init() {
	logLevel = config.GetUint("LOG_LEVEL")
	stackDriverEnabled = config.GetBool("STACKDRIVER_ENABLED")
	gRPCConnectTimeout = config.GetMilliseconds("GRPC_CONNECT_TIMEOUT_MS")
	projectID = config.GetString("PROJECT_ID")
}

// Initialize initializes the logger module.
func Initialize(ctx context.Context) {
	// Do not setup StackDriver client if not configured.
	if !stackDriverEnabled {
		return
	}

	// Setup timeout context for connecting to StackDriver.
	timeoutCtx, cancel := context.WithTimeout(ctx, gRPCConnectTimeout)
	defer cancel()

	// Create StackDriver logger client.
	var err error
	stackDriverClient, err = logging.NewClient(timeoutCtx, projectID)
	if err != nil {
		panic(err)
	}

	// Check StackDriver connection.
	if err = stackDriverClient.Ping(timeoutCtx); err != nil {
		panic(err)
	}

	// Create StackDriver logger instance.
	StackDriverLogger = &Logger{stackDriverClient.Logger(global.ServiceName)}
}

// Finalize finalizes the logging module.
func Finalize() {
	// Check if client and logger are valid.
	if stackDriverClient == nil || StackDriverLogger == nil {
		return
	}

	// Flush logs and properly close logging service connection.
	if err := stackDriverClient.Close(); err != nil {
		now := float64(time.Now().UnixNano()) / float64(time.Second)
		fmt.Fprintf(os.Stderr,
			"\r\x1b[100m%f\x1b[m %s\x1b[m \x1b[100m%12s\x1b[m %s\n",
			now, logLabels[logLevelError], global.ServiceName, err.Error())
	}
}

// Critical logs a message of critical severity.
func Critical(requestCtx context.Context, format string, args ...interface{}) {
	logWithLineNumber(requestCtx, logLevelCritical, format, args...)
}

// Error logs a message of error severity.
func Error(requestCtx context.Context, format string, args ...interface{}) {
	logWithLineNumber(requestCtx, logLevelError, format, args...)
}

// Warn logs a message of warning severity.
func Warn(requestCtx context.Context, format string, args ...interface{}) {
	log(requestCtx, logLevelWarn, format, args...)
}

// Info logs a message of informational severity.
func Info(requestCtx context.Context, format string, args ...interface{}) {
	log(requestCtx, logLevelInfo, format, args...)
}

// Debug logs a message of debugging severity.
func Debug(requestCtx context.Context, format string, args ...interface{}) {
	log(requestCtx, logLevelDebug, format, args...)
}

// log is the general logging utility function used by all log levels.
func log(requestCtx context.Context, level uint, format string, args ...interface{}) {
	// Perform logging only if configured above and within valid log level.
	if level <= logLevelFirst || level >= logLevelLast || level > logLevel {
		return
	}

	// Compose log message.
	message := fmt.Sprintf(format, args...)
	userID, _ := requestCtx.Value(ContextKeyUserID).(string)
	requestID, _ := requestCtx.Value(ContextKeyRequestID).(string)
	if len(requestID) <= 0 {
		requestID = global.ServiceName
	}

	// Log to StackDriver logging service
	if stackDriverClient != nil && StackDriverLogger != nil {
		StackDriverLogger.Log(logging.Entry{
			Severity: logSeverities[level],
			Payload:  message,
			Labels: map[string]string{
				"request_id": requestID,
				"user_id":    userID,
			},
		})
	}

	// now is the current Unix timestamp in floating point.
	now := float64(time.Now().UnixNano()) / float64(time.Second)

	// Reset terminal color.
	fmt.Print("\x1b[m")

	// Log to standard output.
	fmt.Fprintf(os.Stdout,
		"\r\x1b[100m%f\x1b[m %s\x1b[m \x1b[100m%12s\x1b[m %s\n",
		now, logLabels[level], requestID, message)
}

// logWithLineNumber performs usual logging but with an extra line number arg.
func logWithLineNumber(requestCtx context.Context, level uint, format string,
	args ...interface{}) {
	// Get caller file name and line number.
	_, filepath, line, ok := runtime.Caller(2)
	if ok {
		filename := path.Base(filepath)
		format = fmt.Sprintf("%s (%s:%d)", format, filename, line)
	}
	log(requestCtx, level, format, args...)
}
