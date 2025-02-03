package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

// инициализация zap логгера (уровень логгирования INFO)
func Initialize(FileLog bool) {
	var infile zapcore.WriteSyncer
	var core zapcore.Core
	var file *os.File
	var err error

	out := zapcore.AddSync(os.Stdout)
	if FileLog {
		file, err = os.OpenFile("logger.log", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println(err)
			FileLog = false
		} else {
			infile = zapcore.AddSync(file)
		}
	}

	// создаём новую конфигурацию логера
	ConsoleCfg := zap.NewDevelopmentEncoderConfig()
	FileCfg := zap.NewProductionEncoderConfig()
	// устанавливаем время
	ConsoleCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC1123)
	FileCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC1123)
	// устанавливаем уровень
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	// включаем цветовую индикацию для консоли
	ConsoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(ConsoleCfg)
	fileEncoder := zapcore.NewJSONEncoder(FileCfg)

	if FileLog {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, out, level),
			zapcore.NewCore(fileEncoder, infile, level),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, out, level),
		)
	}
	// установка синглтона
	Log = zap.New(core)
}

// // middleware обработчик для zap логгера с логированием полученных и отправленных запросов
// func MiddlewareLogger(h http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
// 		start := time.Now()
// 		responseData := &responseData{
// 			status: 0,
// 			size:   0,
// 		}
// 		reslog := loggingResponseWriter{
// 			ResponseWriter: res, //встраиваем оригинальный http.ResponseWriter
// 			responseData:   responseData,
// 		}
// 		h(&reslog, req)
// 		duration := time.Since(start)
// 		Log.Info("got incoming HTTP request",
// 			zap.String("URI", req.RequestURI),
// 			zap.String("method", req.Method),
// 		)
// 		Log.Info("responce on request",
// 			zap.Int("status", responseData.status),
// 			zap.Int("size", responseData.size),
// 			zap.Duration("duration", duration),
// 		)
// 	})
// }

// // переопределение методов write и WriteHeader для удобного использования middleware
// type (
// 	//структура для хранения сведений об ответе
// 	responseData struct {
// 		status int
// 		size   int
// 	}

// 	//реализация http.ResponseWriter
// 	loggingResponseWriter struct {
// 		http.ResponseWriter //встраиваем оригинальный http.ResponseWriter
// 		responseData        *responseData
// 	}
// )

// func (r *loggingResponseWriter) Write(b []byte) (int, error) {
// 	//запись ответа, используя оригинальный http.ResponseWriter
// 	size, err := r.ResponseWriter.Write(b)
// 	r.responseData.size += size
// 	return size, err
// }

// func (r *loggingResponseWriter) WriteHeader(statusCode int) {
// 	//запись кода статуса, используя оригинальный http.ResponseWriter
// 	r.ResponseWriter.WriteHeader(statusCode)
// 	r.responseData.status = statusCode //код статуса
// }
