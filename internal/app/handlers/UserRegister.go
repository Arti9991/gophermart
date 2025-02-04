package handlers

import (
	"encoding/json"
	"gophermart/internal/app/middleware"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"go.uber.org/zap"
)

// хэндлер для регистрации нового пользователя в системе и выдачи cookie
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
		// декодирование данных пользователя из запроса
		err := json.NewDecoder(req.Body).Decode(&UserData)
		if err != nil {
			logger.Log.Info("Bad request body", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		UserID := models.RandomString(16)
		// хэширование пароля для записи в базу
		CdPassword := CodePassword(UserData.Password)
		// сохранение пользователя в базе данных
		err = hd.StorFunc.SaveUser(UserData.Login, CdPassword, UserID)
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
		// создание JWT для выдачи зарегестрированному пользователю
		JWTstring, err := middleware.BuildJWTString(UserID)
		if err != nil {
			logger.Log.Error("Error in BuildJWTString!", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name:  "Token",
			Value: JWTstring,
		}
		// выдача куки с валидным cookie для сессии данного пользователя
		http.SetCookie(res, cookie)
		res.WriteHeader(http.StatusOK)
	}

}
