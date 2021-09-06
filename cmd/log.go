package cmd

import (
	"os"

	"github.com/go-logr/zapr"
	"github.com/spf13/pflag"
	log "github.com/yittg/golog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logOpt struct {
	Level int8
}

var opt = logOpt{}

func SetLogger() func() {
	logLevel := zap.NewAtomicLevelAt(zapcore.Level(-opt.Level))
	var outputs []zapcore.Core
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleOutput := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		logLevel,
	)
	outputs = append(outputs, consoleOutput)
	zapLog := zap.New(zapcore.NewTee(outputs...))
	log.SetLogger(zapr.NewLogger(zapLog))
	return func() {
		_ = zapLog.Sync()
	}
}

func BindLogFlags(fs *pflag.FlagSet) {
	fs.Int8VarP(&opt.Level, "verbose", "v", opt.Level, "Use to set verbose log level")
}
