package middleware

import (
	"context"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

var UserSession = "userID"

func MiddlewareAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var UserExist bool
		var UserID string

		cookie, err := req.Cookie(UserSession)
		if err != nil {
			UserExist = false
		} else {
			UserID, err = GetUserID(cookie.Value)
			if err != nil {
				logger.Log.Info("Error in JWT", zap.Error(err))
				UserExist = false
			} else {
				UserExist = true
			}
		}

		if !UserExist && !IsLogRegPath(req.RequestURI) {
			logger.Log.Info("User do not exist", zap.String("path", req.RequestURI))
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		//fmt.Printf("\n\nUserID in context: %s\n\n", UserID)
		ctx := context.WithValue(req.Context(), models.CtxKey, models.UserInfo{UserID: UserID, Session: UserExist})
		req = req.WithContext(ctx)
		// передаём управление хендлеру
		h.ServeHTTP(res, req)
	})
}

func IsLogRegPath(path string) bool {
	if strings.Contains(path, "register") || strings.Contains(path, "login") {
		return true
	} else {
		return false
	}
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(UserID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: UserID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", models.ErrorBadToken
	}

	return claims.UserID, nil
}
