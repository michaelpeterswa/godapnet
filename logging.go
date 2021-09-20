package godapnet

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	version := "v0.0.6"
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	logger.Info("godapnet is initializing...", zap.String("version", version))
}
