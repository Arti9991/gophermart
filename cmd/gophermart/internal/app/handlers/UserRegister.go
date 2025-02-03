package handlers

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"go.uber.org/zap"
)

// хэндлер для регистрации нового пользователя в системе
func UserRegister(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			logger.Log.Error("Only POST requests are allowed with this path!", zap.String("method", req.Method))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Header.Get("content-type") != "application/json" {
			logger.Log.Error("Bad content-type header with this path!", zap.String("header", req.Header.Get("content-type")))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		var UserData models.UserRigisterLogin

		err := json.NewDecoder(req.Body).Decode(&UserData)
		if err != nil {
			logger.Log.Info("Bad request body", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		UserID := models.RandomString(16)
		fmt.Println(UserData, UserID)
		err = hd.StorFunc.SaveUser(UserData.Login, UserData.Password, UserID)
		if err != nil {
			if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
				res.WriteHeader(http.StatusConflict)
				return
			} else {
				logger.Log.Error("Error in SaveUser!", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Test Answer UserRegister"))
	}

}

// curl -v -X POST -H "Content-Type: application/json" -d '[
// {"Login":"ID","password":"12345678"}]' http://localhost:8082/api/user/register
