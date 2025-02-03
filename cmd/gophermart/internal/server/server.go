package server

import (
	"gophermart/internal/app/handlers"
	"gophermart/internal/config"
	"gophermart/internal/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var InFileLog = true

type Server struct {
	Config config.Config
	hd     *handlers.HandlersData
}

func InitServer() Server {
	var server Server

	logger.Initialize(InFileLog)
	logger.Log.Info("Logger initialyzed!",
		zap.Bool("In file mode:", InFileLog),
	)

	server.Config = config.InitConf()

	handlers.HandlersDataInit(server.Config.HostAddr, server.Config.AccurAddr)

	return server
}

// создание роутера chi для хэндлеров
func (s *Server) MainRouter() chi.Router {

	rt := chi.NewRouter()
	rt.Route("/api/user/", func(rt chi.Router) {
		rt.Post("/register", handlers.UserRegister(s.hd))
		rt.Post("/login", handlers.UserLogin(s.hd))
	})

	return rt
}

func RunServer() error {
	server := InitServer()

	logger.Log.Info("New server initialyzed!",
		zap.String("Server addres:", server.Config.HostAddr),
		zap.String("DB info:", server.Config.DBAdr),
		zap.String("Accur address:", server.Config.AccurAddr),
	)

	err := http.ListenAndServe(server.Config.HostAddr, server.MainRouter())
	if err != nil {
		logger.Log.Error("New server initialyzed!", zap.Error(err))
		return err
	}
	return nil
}
