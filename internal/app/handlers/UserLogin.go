package handlers

import (
	"encoding/json"
	"gophermart/internal/app/middleware"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// хэндлер для авторизации пользователя и выдачи cookie
func UserLogin(hd *HandlersData) http.HandlerFunc {
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
		// получение хэшированного пароля и идентификатора пользователя из базы по логину
		UserID, passwordCD, err := hd.StorUser.GetUserID(UserData.Login)
		if err != nil {
			if strings.Contains(err.Error(), models.ErrorNoUserString) {
				res.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				logger.Log.Error("Error in GetUserID!", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		// хэширование пароля из запроса
		passwordIN := CodePassword(UserData.Password)
		// проверка подлинности пароля и его соответствие с паролем в базе
		if passwordIN != passwordCD {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		// создание JWT для выдачи авторизованному пользователю
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
