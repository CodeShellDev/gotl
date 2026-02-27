package logger

import (
	"bytes"
	"log"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/ioutils"
	"go.uber.org/zap/zapcore"
)

var stdLoggers map[int]*StdLogger
var stdLoggerIndex int

type StdLogger struct {
	logger *Logger
	levelLoggers map[int]*log.Logger
}

func NewStdLogger(level string, options Options) *StdLogger {
	logger, _ := New(level, options)

	stdLoggerIndex++

	std := &StdLogger{
		logger: logger,
		levelLoggers: createStdLoggers(stdLoggerIndex),
	}

	stdLoggers[stdLoggerIndex] = std

	return std
}

func NewStdLoggerWithDefaults(level string) *StdLogger {
	options := DefaultOptions()
	options.StackDepth++

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

		parts := strings.SplitAfterN(msg, ";", 2)

		if len(parts) != 2 {
			return
		}

		index := parts[0]
		lvl := parts[1]

		i, _ := strconv.Atoi(index)
		level, _ := strconv.Atoi(lvl)
		msg = parts[1]

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

func (std *StdLogger) GetFatal() *log.Logger {
	return std.getStdLevelLogger(zapcore.PanicLevel)
}

func (std *StdLogger) GetError() *log.Logger {
	return std.getStdLevelLogger(zapcore.ErrorLevel)
}

func (std *StdLogger) GetWarn() *log.Logger {
	return std.getStdLevelLogger(zapcore.WarnLevel)
}

func (std *StdLogger) GetInfo() *log.Logger {
	return std.getStdLevelLogger(zapcore.InfoLevel)
}

func (std *StdLogger) GetDebug() *log.Logger {
	return std.getStdLevelLogger(zapcore.DebugLevel)
}

func (std *StdLogger) GetDev() *log.Logger {
	return std.getStdLevelLogger(DeveloperLevel)
}