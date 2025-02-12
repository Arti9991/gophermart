package handlers

import (
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"io"
	"net/http"
	"strconv"

	"github.com/theplant/luhn"
	"go.uber.org/zap"
)

// хэндлер для размещения заказа в системе
func PostOrder(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("content-type") != "text/plain" {
			logger.Log.Error("Bad content-type header with this path!", zap.String("header", req.Header.Get("content-type")))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// чтение тела запроса
		body, err := io.ReadAll(req.Body)
		if err != nil || string(body) == "" {
			logger.Log.Error("Bad request body. ReadAll", zap.Error(err), zap.String("body", string(body)))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// декодируем номер из тела запроса
		numberStr := string(body)
		numberInt, err := strconv.Atoi(string(body))
		if err != nil {
			logger.Log.Error("Bad request body. Atoi", zap.Error(err), zap.Int("number", numberInt))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// // проверяем номер по алгоритму Луна, если отрицательно то ставим статус 422
		if !luhn.Valid(numberInt) {
			logger.Log.Error("Wrong number for luhn", zap.Error(err), zap.Int("number", numberInt))
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		// // получаем информацию о пользователе из контекста, переданного из middleware
		UserInfo := req.Context().Value(models.CtxKey).(models.UserInfo)
		UserID := UserInfo.UserID
		// сохраняем заказ пользователя в базе заказов
		err = hd.StorOrder.SaveNewOrder(UserID, numberStr)
		if err != nil {
			// проверяем ошибку, наличие заказа у этого пользователя или у другого
			// выставляем статус в соответствии с ошикбкой
			if err == models.ErrorUserAlreadyHas {
				logger.Log.Info("User already has that number", zap.Int("number", numberInt), zap.String("UserID", UserID))
				res.WriteHeader(http.StatusOK)
				return
			} else if err == models.ErrorAnotherUserHas {
				logger.Log.Info("Another user has that number", zap.Int("number", numberInt))
				res.WriteHeader(http.StatusConflict)
				return
			} else {
				logger.Log.Error("Error in SaveNewOrder", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusAccepted)
	}

}
