package server

import (
	"gophermart/internal/app/handlers"
	"gophermart/internal/config"
	"gophermart/internal/logger"
	"gophermart/internal/storage/database"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

var InFileLog = true

type Server struct {
	Config   config.Config
	hd       *handlers.HandlersData
	DataBase *database.DBStor
}

func InitServer() Server {
	var server Server
	var err error

	// установка сида для случайных чисел
	rand.Seed(uint64(time.Now().UnixNano()))

	logger.Initialize(InFileLog)
	logger.Log.Info("Logger initialyzed!",
		zap.Bool("In file mode:", InFileLog),
	)

	server.Config = config.InitConf()
	server.DataBase, err = database.DBinit(server.Config.DBAdr)
	if err != nil {
		logger.Log.Fatal("Error in initialyzed database", zap.Error(err))
	}
	server.hd = handlers.HandlersDataInit(server.Config.HostAddr, server.Config.AccurAddr, server.DataBase)
	return server
}

// создание роутера chi для хэндлеров
func (s *Server) MainRouter() chi.Router {

	rt := chi.NewRouter()
	rt.Use(logger.MiddlewareLogger)
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
