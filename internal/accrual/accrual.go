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
	AccAddr   string
	RetryTime time.Duration
}

func AccrualDataInit(AccrualAddr string) *AccrualData {
	return &AccrualData{AccAddr: AccrualAddr + "/api/orders/", RetryTime: 0}
}

// функция периодически отправляющая данные к сервису расчета количеств баллов
func (AccDt *AccrualData) LoadNumberToAPI(numCh <-chan models.OrderAns) chan models.OrderAns {
	outCh := make(chan models.OrderAns)
	go func() {
		for outBuff := range numCh {
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
					response.Body.Close()
					continue
				}
				response.Body.Close()
				time.Sleep(timeSleep)
				continue
			} else if response.StatusCode == http.StatusNoContent {
				outBuff.Accrual = 0.0
				outBuff.Status = "EMPTY"
			} else if response.StatusCode == http.StatusOK {
				// читаем поток из тела ответа
				err = json.NewDecoder(response.Body).Decode(&outBuff)
				if err != nil && err != io.EOF {
					logger.Log.Error("Error in Decode", zap.Error(err))
					response.Body.Close()
					continue
				}
			}
			// хаписываем результат в канал
			outCh <- outBuff
			response.Body.Close()
			time.Sleep(50 * time.Millisecond)

		}
	}()
	return outCh
}
