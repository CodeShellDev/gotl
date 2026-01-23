package logger

import (
	"go.uber.org/zap/zapcore"
)

func (logger *Logger) Dev(data ...any) {
	if logger.IsDev() {
		logger.zap.Log(DeveloperLevel, format(data...))
	}
}

func (logger *Logger) Debug(data ...any) {
	if logger.IsDebug() {
		logger.zap.Debug(format(data...))
	}
}

func (logger *Logger) Info(data ...any) {
	if logger.IsInfo() {
		logger.zap.Info(format(data...))
	}
}

func (logger *Logger) Warn(data ...any) {
	if logger.IsWarn() {
		logger.zap.Warn(format(data...))
	}
}

func (logger *Logger) Error(data ...any) {
	if logger.IsError() {
		logger.zap.Error(format(data...))
	}
}

func (logger *Logger) Fatal(data ...any) {
	if logger.IsFatal() {
		logger.zap.Fatal(format(data...))
	}
}


func (logger *Logger) IsDev() bool {
	return logger.level.Level().Enabled(DeveloperLevel)
}

func (logger *Logger) IsDebug() bool {
	return logger.level.Level().Enabled(zapcore.DebugLevel)
}

func (logger *Logger) IsInfo() bool {
	return logger.level.Level().Enabled(zapcore.InfoLevel)
}

func (logger *Logger) IsWarn() bool {
	return logger.level.Level().Enabled(zapcore.WarnLevel)
}

func (logger *Logger) IsError() bool {
	return logger.level.Level().Enabled(zapcore.ErrorLevel)
}

func (logger *Logger) IsFatal() bool {
	return logger.level.Level().Enabled(zapcore.FatalLevel)
}

func Dev(data ...any) {
	defaultLogger.Dev(data...)
}

func Debug(data ...any) {
	defaultLogger.Debug(data...)
}

func Info(data ...any) {
	defaultLogger.Info(data...)
}

func Warn(data ...any) {
	defaultLogger.Warn(data...)
}

func Error(data ...any) {
	defaultLogger.Error(data...)
}

func Fatal(data ...any) {
	defaultLogger.Fatal(data...)
}


func IsDev() bool {
	return defaultLogger.IsDev()
}

func IsDebug() bool {
	return defaultLogger.IsDebug()
}

func IsInfo() bool {
	return defaultLogger.IsInfo()
}

func IsWarn() bool {
	return defaultLogger.IsWarn()
}

func IsError() bool {
	return defaultLogger.IsError()
}

func IsFatal() bool {
	return defaultLogger.IsFatal()
}