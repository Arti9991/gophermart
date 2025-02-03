package handlers

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// хэндлер для получения оригинального URL по укороченному
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

		err := json.NewDecoder(req.Body).Decode(&UserData)
		if err != nil {
			logger.Log.Info("Bad request body", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		UserID, password, err := hd.StorFunc.GetUserID(UserData.Login)
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
		fmt.Println(UserID, password)
		if password != UserData.Password {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Test Answer UserRegister"))
	}

}
