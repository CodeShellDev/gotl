package logger

import (
	"bytes"
	"log"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/ioutils"
	"go.uber.org/zap/zapcore"
)
func getPrefixFromLevel(level zapcore.Level) string {
	return strconv.Itoa(int(level)) + ";"
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

		parts := strings.SplitAfterN(msg, ";", 1)

		if len(parts) == 0 {
			return
		}

		prefix := parts[0]

		level, _ := strconv.Atoi(prefix)
		msg = parts[1]

		msg = normalizeMessage(msg)

		switch (level) {
		case int(zapcore.FatalLevel):
			Fatal(msg)
		case int(zapcore.ErrorLevel):
			Error(msg)
		case int(zapcore.WarnLevel):
			Warn(msg)
		case int(zapcore.InfoLevel):
			Info(msg)
		case int(zapcore.DebugLevel):
			Debug(msg)
		case int(DeveloperLevel):
			Dev(msg)
		default:
			Info(msg)
		}
	},
}

var FatalLog *log.Logger = log.New(writer, getPrefixFromLevel(zapcore.FatalLevel), 0)
var ErrorLog *log.Logger = log.New(writer, getPrefixFromLevel(zapcore.ErrorLevel), 0)
var WarnLog *log.Logger = log.New(writer, getPrefixFromLevel(zapcore.WarnLevel), 0)

var InfoLog *log.Logger = log.New(writer, getPrefixFromLevel(zapcore.InfoLevel), 0)
var DebugLog *log.Logger = log.New(writer, getPrefixFromLevel(zapcore.DebugLevel), 0)
var DevLog *log.Logger = log.New(writer, getPrefixFromLevel(DeveloperLevel), 0)
