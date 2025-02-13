package handlers

import (
	"encoding/json"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strconv"

	"github.com/theplant/luhn"
	"go.uber.org/zap"
)

// хэндлер для размещения заказа в системе
func WithdrawOrder(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("content-type") != "application/json" {
			logger.Log.Error("Bad content-type header with this path!", zap.String("header", req.Header.Get("content-type")))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		var WithData models.WithData
		// декодирование данных пользователя из запроса
		err := json.NewDecoder(req.Body).Decode(&WithData)
		if err != nil {
			logger.Log.Info("Bad request body", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		numberInt, err := strconv.Atoi(WithData.Number)
		if err != nil {
			logger.Log.Error("Bad request body. Atoi", zap.Error(err), zap.Int("number", numberInt))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// проверяем номер по алгоритму Луна, если отрицательно то ставим статус 422
		if !luhn.Valid(numberInt) {
			logger.Log.Error("Wrong number for luhn", zap.Error(err), zap.Int("number", numberInt))
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		// получаем информацию о пользователе из контекста, переданного из middleware
		UserInfo := req.Context().Value(models.CtxKey).(models.UserInfo)
		UserID := UserInfo.UserID

		err = hd.StorOrder.SaveWithdrawOrder(UserID, WithData.Number, WithData.Sum)
		if err != nil {
			if err == models.ErrorNoSuchBalance {
				logger.Log.Info("Not enough money on balance")
				res.WriteHeader(http.StatusPaymentRequired)
				return
			} else if err == models.ErrorAlreadyTakenNumber {
				logger.Log.Error("Bad order number", zap.Error(err))
				res.WriteHeader(http.StatusUnprocessableEntity)
				return
			} else {
				logger.Log.Error("Error in MinusUserBalance", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		res.WriteHeader(http.StatusOK)
	}

}
