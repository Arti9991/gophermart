package server

import (
	"gophermart/internal/config"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

var InFileLog = true

type Server struct {
	Config config.Config
}

func InitServer() Server {
	var server Server

	logger.Initialize(InFileLog)
	logger.Log.Info("Logger initialyzed!",
		zap.Bool("In file mode:", InFileLog),
	)

	server.Config = config.InitConf()

	return server
}

func RunServer() error {
	server := InitServer()

	logger.Log.Info("New server initialyzed!",
		zap.String("Server addres:", server.Config.HostAddr),
		zap.String("DB info:", server.Config.DBAdr),
		zap.String("Accur address:", server.Config.AccurAddr),
	)

	return nil
}
