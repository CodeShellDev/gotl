package logger

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *Logger

type Logger struct {
	zap			*zap.Logger
	level		*zap.AtomicLevel
	options		Options
	transform	func(content string) string
}

type Options struct {
	EncodeLevel 	zapcore.LevelEncoder
	EncodeCaller 	zapcore.CallerEncoder
	EncodeDuration 	zapcore.DurationEncoder
	EncodeTime 		zapcore.TimeEncoder
	StackDepth 		int
}

func DefaultOptions() Options {
	return Options{
		EncodeTime: zapcore.TimeEncoderOfLayout("02.01 15:04"),
		EncodeLevel: customEncodeLevel,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		StackDepth: 1,
	}
}

func NewWithDefaults(level string) (*Logger, error) {
	return New(level, DefaultOptions())
}

func New(level string, options Options) (*Logger, error) {
	lvl := parseLevel(strings.ToLower(level))
	atomicLevel := zap.NewAtomicLevelAt(lvl)

	cfg := zap.Config{
		Level:       atomicLevel,
		Encoding:    "console",
		OutputPaths: []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    options.EncodeLevel,
			EncodeTime:     options.EncodeTime,
			EncodeDuration: options.EncodeDuration,
			EncodeCaller:   modifyCaller(options.EncodeCaller),
		},
	}

	z, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(options.StackDepth),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{
		zap:     z,
		level:   &atomicLevel,
		options: options,
	}, nil
}

func modifyCaller(encoder zapcore.CallerEncoder) zapcore.CallerEncoder {
	return func(caller zapcore.EntryCaller, pae zapcore.PrimitiveArrayEncoder) {
		path := caller.File

		i := strings.Index(path, "@")
		if i != -1 {
			// find and remove @X.Y.Z
			slashI := strings.Index(path[i:], "/")
			if slashI != -1 {
				path = path[:i] + path[i+slashI:]
			}
		}

		caller.File = path

		encoder(caller, pae)
	}
}

func format(data ...any) string {
	res := ""

	for _, item := range data {
		switch value := item.(type) {
		case string:
			res += value
		case []byte:
			res += string(value)
		case int:
			res += strconv.Itoa(value)
		case int32:
			res += strconv.Itoa(int(value))
		case int64:
			res += strconv.Itoa(int(value))
		case bool:
			if value {
				res += "true"
			} else {
				res += "false"
			}
		default:
			lines := strings.Split(jsonutils.Pretty(value), "\n")

			lineStr := ""

			for _, line := range lines {
				lineStr += "\n" + startColor(color.RGBA{ R: 0, G: 135, B: 95,}) + line + resetColor()
			}
			res += lineStr
		}
	}

	return res
}

func transform(content string, fn func(content string) string) string {
	if fn != nil {
		return fn(content)
	}

	return content
}

func Init(level string) error {
	l, err := NewWithDefaults(level)
	if err != nil {
		return err
	}

	defaultLogger = l

	return nil
}

func InitWith(level string, options Options) error {
	l, err := New(level, options)
	if err != nil {
		return err
	}
	
	defaultLogger = l

	return nil
}

func Level() string {
	if defaultLogger != nil {
		return defaultLogger.Level()
	}

	return ""
}

func Sync() {
	if defaultLogger != nil {
		defaultLogger.Sync()
	}
}

func Get() *Logger {
	return defaultLogger
}

func (logger *Logger) Level() string {
	return levelString(logger.level.Level())
}

func (logger *Logger) SetLevel(level string) {
	logger.level.SetLevel(parseLevel(strings.ToLower(level)))
}

func (logger *Logger) SetTransform(transform func(content string) string) {
	logger.transform = transform
}

func (logger *Logger) Clone() (*Logger, error) {
	return New(logger.level.Level().String(), logger.options)
}

func (logger *Logger) Sub(level string) *Logger {
	atomicLevel := zap.NewAtomicLevelAt(parseLevel(strings.ToLower(level)))

	return &Logger{
		zap:     logger.zap,
		level:   &atomicLevel,
		options: logger.options,
	}
}

func (logger *Logger) Sync() {
	_ = logger.zap.Sync()
}