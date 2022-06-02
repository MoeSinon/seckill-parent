package trigger

import (
	"go.uber.org/zap"
)

type Pos struct {
	Position int
	length   int
}

func Zaplog() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	return logger.Sugar()
}
