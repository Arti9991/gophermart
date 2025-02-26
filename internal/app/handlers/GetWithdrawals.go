package handlers

import (
	"encoding/json"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"

	"go.uber.org/zap"
)

// хэндлер для размещения заказа в системе (заглушка)
func GetWithrawals(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		UserInfo := req.Context().Value(models.CtxKey).(models.UserInfo)
		UserID := UserInfo.UserID

		WithdrawList, err := hd.StorOrder.GetUserWithdrawals(UserID)
		if err != nil {
			if err == models.ErrorNoOrdersUser {
				logger.Log.Info("For this user:", zap.Error(err), zap.String("UserID", UserID))
				res.WriteHeader(http.StatusNoContent)
				return
			}
			logger.Log.Error("Error in get orders", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		// кодирование тела ответа
		out, err := json.Marshal(WithdrawList)
		if err != nil {
			logger.Log.Error("Bad Marshall for out body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(out)
	}

}
