package main

import (
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logCfg := logger.Config{
		LogFile:   "app.log",
		LogLevel:  "debug",
		AppName:   "form-service",
		AddCaller: true,
	}

	if err := logger.Init(logCfg); err != nil {
		panic(err)
	}

	defer logger.Sync()

	logger := logger.Get()

	cfg, err := config.Init(".env")
	if err != nil {
		logger.Error("error init config",
			zap.String("path", ".env"),
			zap.Error(err))
		return
	}

}
