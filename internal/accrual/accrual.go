package accrual

import (
	"encoding/json"
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type AccrualData struct {
	AccAddr     string
	waitCh      chan time.Duration
	NumberWorks int
}

func AccrualDataInit(AccrualAddr string, numberWorkers int) *AccrualData {
	durCh := make(chan time.Duration)
	return &AccrualData{AccAddr: AccrualAddr + "api/orders/", waitCh: durCh, NumberWorks: numberWorkers}
}

// функция периодически отправляющая данные к сервису расчета количеств баллов
func (AccDt *AccrualData) LoadNumberToApi(RequestPool chan models.OrderAns, ResulReqCh chan models.OrderAns) {
	go func() {
		for {
			select {
			case outBuff := <-RequestPool:
				// формируем URI для запроса с номером заказа
				reqURI := AccDt.AccAddr + outBuff.Number
				// выполняем запрос для получения информации по рассчету баллов
				response, err := http.Get(reqURI)
				if err != nil {
					logger.Log.Error("Error in GET", zap.Error(err))
					continue
				}
				// если превышен лимит запросов, то парсим хэдер и засыпаем на положенное время
				if response.StatusCode == http.StatusTooManyRequests {
					sleepTimeStr := response.Header.Get("Retry-After")
					timeSleep, err2 := time.ParseDuration(sleepTimeStr)
					if err2 != nil {
						logger.Log.Error("Error in ParseDuration", zap.Error(err))
						continue
					}
					AccDt.waitCh <- timeSleep
					time.Sleep(timeSleep)
				} else if response.StatusCode == http.StatusNoContent {
					outBuff.Accrual = 0.0
					outBuff.Status = "NEW"
					ResulReqCh <- outBuff
				} else if response.StatusCode == http.StatusOK {
					// читаем поток из тела ответа
					err = json.NewDecoder(response.Body).Decode(&outBuff)
					if err != nil && err != io.EOF {
						logger.Log.Error("Error in Decode", zap.Error(err))
					}
					ResulReqCh <- outBuff
				}
			case timeSleep := <-AccDt.waitCh:
				time.Sleep(timeSleep)
			default:
				continue
			}
		}
	}()
}
