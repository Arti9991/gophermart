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
		file, err = os.OpenFile("logger.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
