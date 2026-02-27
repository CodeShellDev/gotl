package logger

import (
	"bytes"
	"log"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/ioutils"
	"go.uber.org/zap/zapcore"
)

var stdLoggers = map[int]*StdLogger{}
var stdLoggerIndex int

var defaultSdtLogger *StdLogger

type StdLogger struct {
	logger *Logger
	levelLoggers map[int]*log.Logger
}

func InitStdLogger(level string) {
	defaultSdtLogger = NewStdLoggerWithDefaults(level)
}

func InitStdLoggerWith(level string, options Options) {
	defaultSdtLogger = NewStdLogger(level, options)
}

func NewStdLogger(level string, options Options) *StdLogger {
	options.StackDepth += 2
	logger, _ := New(level, options)

	std := &StdLogger{
		logger: logger,
		levelLoggers: createStdLoggers(stdLoggerIndex),
	}

	stdLoggers[stdLoggerIndex] = std

	stdLoggerIndex++

	return std
}

func NewStdLoggerWithDefaults(level string) *StdLogger {
	options := DefaultOptions()

	return NewStdLogger(level, options)
}

func getStdLoggerByIndex(i int) *StdLogger {
	return stdLoggers[i]
}

func addStdLevelLoggerToLoggers(loggers map[int]*log.Logger, i int, level zapcore.Level) {
	loggers[int(level)] = log.New(writer, encodeDataForStdLogger(i, level), 0)
}

func encodeDataForStdLogger(i int, level zapcore.Level) string {
	return strconv.Itoa(i) + ";" + strconv.Itoa(int(level)) + ";"
}

func normalizeMessage(msg string) string {
	msg = strings.TrimSuffix(msg, "\n")

	msg = strings.ToUpper(msg[:1]) + msg[1:]

	return msg
}

var writer = &ioutils.InterceptWriter{
	Writer: &bytes.Buffer{},
	Hook: func(bytes []byte) {
		msg := string(bytes)
		if len(msg) == 0 {
			return
		}

		parts := strings.SplitAfterN(msg, ";", 3)

		if len(parts) != 3 {
			return
		}

		index := parts[0]
		lvl := parts[1]

		i, _ := strconv.Atoi(index)
		level, _ := strconv.Atoi(lvl)
		msg = parts[2]

		msg = normalizeMessage(msg)

		switch (level) {
		case int(zapcore.FatalLevel):
			getStdLoggerByIndex(i).logger.Fatal(msg)
		case int(zapcore.ErrorLevel):
			getStdLoggerByIndex(i).logger.Error(msg)
		case int(zapcore.WarnLevel):
			getStdLoggerByIndex(i).logger.Warn(msg)
		case int(zapcore.InfoLevel):
			getStdLoggerByIndex(i).logger.Info(msg)
		case int(zapcore.DebugLevel):
			getStdLoggerByIndex(i).logger.Debug(msg)
		case int(DeveloperLevel):
			getStdLoggerByIndex(i).logger.Dev(msg)
		default:
			getStdLoggerByIndex(i).logger.Info(msg)
		}
	},
}

func createStdLoggers(i int) map[int]*log.Logger {
	loggers := map[int]*log.Logger{}

	addStdLevelLoggerToLoggers(loggers, i, zapcore.PanicLevel)
	addStdLevelLoggerToLoggers(loggers, i, zapcore.ErrorLevel)
	addStdLevelLoggerToLoggers(loggers, i, zapcore.WarnLevel)

	addStdLevelLoggerToLoggers(loggers, i, zapcore.InfoLevel)
	addStdLevelLoggerToLoggers(loggers, i, zapcore.DebugLevel)
	addStdLevelLoggerToLoggers(loggers, i, DeveloperLevel)

	return loggers
}

func (std *StdLogger) getStdLevelLogger(level zapcore.Level) *log.Logger {
	return std.levelLoggers[int(level)]
}

func (std *StdLogger) StdFatal() *log.Logger {
	return std.getStdLevelLogger(zapcore.PanicLevel)
}

func (std *StdLogger) StdError() *log.Logger {
	return std.getStdLevelLogger(zapcore.ErrorLevel)
}

func (std *StdLogger) StdWarn() *log.Logger {
	return std.getStdLevelLogger(zapcore.WarnLevel)
}

func (std *StdLogger) StdInfo() *log.Logger {
	return std.getStdLevelLogger(zapcore.InfoLevel)
}

func (std *StdLogger) StdDebug() *log.Logger {
	return std.getStdLevelLogger(zapcore.DebugLevel)
}

func (std *StdLogger) StdDev() *log.Logger {
	return std.getStdLevelLogger(DeveloperLevel)
}


func StdFatal() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(zapcore.PanicLevel)
}

func StdError() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(zapcore.ErrorLevel)
}

func StdWarn() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(zapcore.WarnLevel)
}

func StdInfo() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(zapcore.InfoLevel)
}

func StdDebug() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(zapcore.DebugLevel)
}

func StdDev() *log.Logger {
	return defaultSdtLogger.getStdLevelLogger(DeveloperLevel)
}