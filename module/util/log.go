package util

import (
	"go.uber.org/zap"
)

func Zaplog() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	return logger.Sugar()
}

// func Exchangelogmod() {

// }
