package accrual

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type AccrualData struct {
	AccAddr        string
	RetryTime      time.Duration
	RequestPool    chan models.OrderAns
	BaseUpdateChan chan models.OrderAns
	waitCh         chan time.Duration
	SemaCh         chan struct{}
}

func AccrualDataInit(AccrualAddr string, numberWorkers int) *AccrualData {
	workers := make(chan models.OrderAns, numberWorkers)
	update := make(chan models.OrderAns)
	sema := make(chan struct{}, numberWorkers)
	return &AccrualData{AccAddr: AccrualAddr + "api/orders/", RetryTime: 0, RequestPool: workers, BaseUpdateChan: update, SemaCh: sema}
}

// функция периодически отправляющая данные к сервису расчета количеств баллов
func (AccDt *AccrualData) LoadNumberToApi(outBuff models.OrderAns) models.OrderAns {
	fmt.Println("In default")
	fmt.Println(outBuff)
	// формируем URI для запроса с номером заказа
	reqURI := AccDt.AccAddr + outBuff.Number
	// выполняем запрос для получения информации по рассчету баллов
	response, err := http.Get(reqURI)
	if err != nil {
		response.Body.Close()
		logger.Log.Error("Error in GET", zap.Error(err))
		return outBuff
	}
	defer response.Body.Close()
	// если превышен лимит запросов, то парсим хэдер и засыпаем на положенное время
	if response.StatusCode == http.StatusTooManyRequests {
		sleepTimeStr := response.Header.Get("Retry-After")
		timeSleep, err2 := time.ParseDuration(sleepTimeStr)
		if err2 != nil {
			logger.Log.Error("Error in ParseDuration", zap.Error(err))
			return outBuff
		}
		AccDt.waitCh <- timeSleep
		time.Sleep(timeSleep)
	} else if response.StatusCode == http.StatusNoContent {
		outBuff.Accrual = 0.0
		outBuff.Status = "INVALID"
	} else if response.StatusCode == http.StatusOK {
		// читаем поток из тела ответа
		err = json.NewDecoder(response.Body).Decode(&outBuff)
		if err != nil && err != io.EOF {
			logger.Log.Error("Error in Decode", zap.Error(err))
			return outBuff
		}
	}
	return outBuff
}
