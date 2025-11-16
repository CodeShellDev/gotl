package logger

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _logLevel = ""

var logger *zap.Logger

type Options struct {
	TimeLayout string
	EncodeLevel zapcore.LevelEncoder
	StackDepth int
}

func Level() string {
	return levelString(logger.Level())
}

func Sync() {
	logger.Sync()
}

// Initialize logger with level string
func Init(level string) error {
	return initialize(level, Options{
		TimeLayout: "02.01 15:04",
		EncodeLevel: customEncodeLevel,
		StackDepth: 1,
	})
}

// Initialize logger with level string and options
func InitWith(level string, options Options) error {
	return initialize(level, options)
}

func initialize(level string, option Options) error {
	_logLevel = strings.ToLower(level)

	logLevel := parseLevel(_logLevel)

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling:    nil,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    option.EncodeLevel,
			EncodeTime:     zapcore.TimeEncoderOfLayout(option.TimeLayout),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	var err error

	logger, err = cfg.Build(zap.AddCaller(), zap.AddCallerSkip(option.StackDepth))

	return err
}

func format(data ...any) string {
	res := ""

	for _, item := range data {
		switch value := item.(type) {
		case string:
			res += value
		case int:
			res += strconv.Itoa(value)
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
				lineStr += "\n" + startColor(color.RGBA{ R: 0, G: 135, B: 95,}) + line + endColor()
			}
			res += lineStr
		}
	}

	return res
}