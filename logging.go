package godapnet

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func InitializeLogger(externalLogger *zap.Logger) {
	logger = externalLogger
	logger.Info("godapnet is initializing...")
}
