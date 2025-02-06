package server

import (
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/middleware"
	"gophermart/internal/config"
	"gophermart/internal/logger"
	"gophermart/internal/storage/database"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

//var key = []byte{99, 65, 113, 122, 87, 106, 113, 81, 114, 115, 66, 117, 107, 81, 116, 108, 73, 77, 75, 111, 89, 71, 79, 106, 118, 76, 69, 106, 115, 116, 75, 101}

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
	server.DataBase, err = database.DBUserInit(server.Config.DBAdr)
	if err != nil {
		logger.Log.Fatal("Error in initialyzed database for users", zap.Error(err))
	}
	err = server.DataBase.DBOrdersInit()
	if err != nil {
		logger.Log.Fatal("Error in initialyzed database for orders", zap.Error(err))
	}
	server.hd = handlers.HandlersDataInit(server.Config.HostAddr, server.Config.AccurAddr, server.DataBase, server.DataBase)
	return server
}

// создание роутера chi для хэндлеров
func (s *Server) MainRouter() chi.Router {

	rt := chi.NewRouter()
	rt.Use(middleware.MiddlewareLogger, middleware.MiddlewareAuth)
	rt.Route("/api/user/", func(rt chi.Router) {
		rt.Post("/register", handlers.UserRegister(s.hd))
		rt.Post("/login", handlers.UserLogin(s.hd))
		rt.Post("/orders", handlers.PostOrder(s.hd))
		rt.Get("/orders", handlers.GetOrders(s.hd))
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
