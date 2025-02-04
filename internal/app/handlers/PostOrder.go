package handlers

import (
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// хэндлер для размещения заказа в системе (заглушка)
func PostOrder(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("content-type") != "text/plain" {
			logger.Log.Error("Bad content-type header with this path!", zap.String("header", req.Header.Get("content-type")))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil || string(body) == "" {
			logger.Log.Info("Bad request body", zap.String("body", string(body)))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		UserInfo := req.Context().Value(models.CtxKey).(models.UserInfo)
		UserID := UserInfo.UserID

		ansStr := "\n" + "user ID is:" + UserID + "\n" + "order is:" + string(body) + "\n"

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(ansStr))
	}

}
