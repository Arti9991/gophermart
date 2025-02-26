package middleware

import (
	"gophermart/internal/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// middleware обработчик для zap логгера с логированием полученных и отправленных запросов
func MiddlewareLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		reslog := loggingResponseWriter{
			ResponseWriter: res, //встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&reslog, req)
		duration := time.Since(start)
		logger.Log.Info("got incoming HTTP request",
			zap.String("URI", req.RequestURI),
			zap.String("method", req.Method),
		)
		logger.Log.Info("responce on request",
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	})
}

// переопределение методов write и WriteHeader для удобного использования middleware
type (
	//структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	//реализация http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter //встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	//запись ответа, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	//запись кода статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode //код статуса
}
