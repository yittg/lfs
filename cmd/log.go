package main

import (
	"github.com/go-logr/zapr"
	log "github.com/yittg/golog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setLogger() {
	if isDevelopment() {
		setDevelopmentLogger()
	} else {
		setProductionLogger()
	}
}

func setDevelopmentLogger() {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(-128))
	zapLog, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	log.SetLogger(zapr.NewLogger(zapLog))
}

func setProductionLogger() {
	zapLog, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	log.SetLogger(zapr.NewLogger(zapLog))
}
