package logger

import (
	"go.uber.org/zap/zapcore"
)

func (logger *Logger) Dev(data ...any) {
	ok := logger.zap.Check(DeveloperLevel, format(data...))

	if ok != nil {
		ok.Write()
	}
}

func (logger *Logger) Debug(data ...any) {
	logger.zap.Debug(format(data...))
}

func (logger *Logger) Info(data ...any) {
	logger.zap.Info(format(data...))
}

func (logger *Logger) Warn(data ...any) {
	logger.zap.Warn(format(data...))
}

func (logger *Logger) Error(data ...any) {
	logger.zap.Error(format(data...))
}

func (logger *Logger) Fatal(data ...any) {
	logger.zap.Fatal(format(data...))
}


func (logger *Logger) IsDev() bool {
	return logger.zap.Level().Enabled(DeveloperLevel)
}

func (logger *Logger) IsDebug() bool {
	return logger.zap.Level().Enabled(zapcore.DebugLevel)
}

func (logger *Logger) IsInfo() bool {
	return logger.zap.Level().Enabled(zapcore.InfoLevel)
}

func (logger *Logger) IsWarn() bool {
	return logger.zap.Level().Enabled(zapcore.WarnLevel)
}

func (logger *Logger) IsError() bool {
	return logger.zap.Level().Enabled(zapcore.ErrorLevel)
}

func (logger *Logger) IsFatal() bool {
	return logger.zap.Level().Enabled(zapcore.FatalLevel)
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