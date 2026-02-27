package logger

import (
	"log"
	"strings"

	"go.uber.org/zap/zapcore"
)

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
		levelLoggers: createStdLoggers(logger),
	}

	return std
}

func NewStdLoggerWithDefaults(level string) *StdLogger {
	return NewStdLogger(level, DefaultOptions())
}

type zapWriter struct {
	logger *Logger
}

func (z *zapWriter) Write(bytes []byte) (n int, err error) {
	msg := strings.TrimSpace(string(bytes))

	if msg == "" {
		return len(bytes), nil
	}

	switch z.logger.level.Level() {
	case zapcore.FatalLevel:
		z.logger.Fatal(msg)
	case zapcore.ErrorLevel:
		z.logger.Error(msg)
	case zapcore.WarnLevel:
		z.logger.Warn(msg)
	case zapcore.InfoLevel:
		z.logger.Info(msg)
	case zapcore.DebugLevel:
		z.logger.Debug(msg)
	default:
		z.logger.Info(msg)
	}

	return len(bytes), nil
}

func createStdLoggers(logger *Logger) map[int]*log.Logger {
	loggers := map[int]*log.Logger{}

	addStdLevelLogger(loggers, logger, zapcore.PanicLevel)
	addStdLevelLogger(loggers, logger, zapcore.ErrorLevel)
	addStdLevelLogger(loggers, logger, zapcore.WarnLevel)
	addStdLevelLogger(loggers, logger, zapcore.InfoLevel)
	addStdLevelLogger(loggers, logger, zapcore.DebugLevel)
	addStdLevelLogger(loggers, logger, DeveloperLevel)

	return loggers
}

func addStdLevelLogger(loggers map[int]*log.Logger, logger *Logger, level zapcore.Level) {
	w := &zapWriter{logger: logger}

	loggers[int(level)] = log.New(w, "", 0)
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