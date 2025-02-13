package database

import (
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var QuerryCreateTypeStatus = `CREATE TYPE status AS ENUM('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED', 'WITHDRAW', 'EMPTY');`
var QuerryCreateorderStor = `
	CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
	user_id VARCHAR(16) NOT NULL,
	number VARCHAR(32) NOT NULL UNIQUE,
    status status NOT NULL,
    accrual NUMERIC(10,2) NOT NULL,
	uploaded_at TIMESTAMPTZ
	);`

var QuerrySaveNewOrder = `INSERT INTO orders (id, user_id, number, status, accrual, uploaded_at)
  	VALUES  (DEFAULT, $1, $2, 'NEW', 0.0, $3);`
var QuerryGetUserForNumber = `SELECT user_id FROM orders 
	WHERE number = $1 LIMIT 1;`
var QuerryGetUserOrders = `SELECT number, status, accrual, uploaded_at FROM orders 
	WHERE user_id = $1 AND (status != 'WITHDRAW' OR status != 'REGISTERED' or status != 'EMPTY') ORDER BY uploaded_at DESC;`
var QuerryGetAccurOrders = `SELECT number, status, user_id  FROM orders
	WHERE status = 'NEW' OR status = 'PROCESSING' OR status = 'REGISTERED';`
var QuerrySaveAccurOrders = `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3;`
var QuerrySaveWithdrawOrder = `INSERT INTO orders (id, user_id, number, status, accrual, uploaded_at)
  	VALUES  (DEFAULT, $1, $2, 'WITHDRAW', $3, $4);`
var QuerryGetUserWithdraw = `SELECT number, accrual, uploaded_at FROM orders 
	WHERE user_id = $1 AND status = 'WITHDRAW' ORDER BY uploaded_at DESC;`

// инициализация хранилища и создание/подключение к таблице
func (db *DBStor) DBOrdersInit() error {
	// создаем таблицу заказов
	_, err := db.DB.Exec(QuerryCreateorderStor)
	if err != nil {
		// если в таблице есть неопределенный тип, определяем его
		if strings.Contains(err.Error(), pgerrcode.UndefinedObject) {
			// определеяем тип для хранения статуса заказа
			_, err = db.DB.Exec(QuerryCreateTypeStatus)
			if err != nil {
				return err
			}
			// создаем таблицу для заказов
			_, err = db.DB.Exec(QuerryCreateorderStor)
			if err != nil {
				return err
			}
			logger.Log.Info("✓ created orders table with new status type!")
			return nil
		} else {
			return err
		}
	}
	logger.Log.Info("✓ created orders table!")
	return nil
}

// функция для сохранения нового заказа пользователя в базу
func (db *DBStor) SaveNewOrder(UserID string, number string) error {
	uploaded := time.Now()
	var err error
	var UserID2 string
	_, err = db.DB.Exec(QuerrySaveNewOrder, UserID, number, uploaded)
	if err != nil {
		// проверяем ошибку от SQL, если это уникальность то проверяем пользователя
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			// получаем пользователя с таким заказом
			row := db.DB.QueryRow(QuerryGetUserForNumber, number)
			err = row.Scan(&UserID2)
			if err != nil {
				return err
			}
			// если это тот же пользователь, что и отправил заказ
			// отправляем соответствующую ошибку
			if UserID2 == UserID {
				return models.ErrorUserAlreadyHas
			} else {
				return models.ErrorAnotherUserHas
			}
		} else {
			return err
		}
	}
	return nil
}

// функция для получения информации о всех заказах пользователя из базы
func (db *DBStor) GetUserOrders(UserID string) (models.UserOrdersList, error) {
	var err error
	var ordersList models.UserOrdersList
	var OpTime time.Time
	var Stat string
	rows, err := db.DB.Query(QuerryGetUserOrders, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.UserOrder
		err := rows.Scan(&order.Number, &Stat, &order.Accrual, &OpTime)
		if err != nil {
			return nil, err
		}
		if Stat == "EMPTY" {
			order.Status = "NEW"
		} else {
			order.Status = Stat
		}
		order.LoadedTime = OpTime.Format(time.RFC3339)
		ordersList = append(ordersList, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ordersList) == 0 {
		return nil, models.ErrorNoOrdersUser
	}
	return ordersList, nil
}

// функция для циклического получения из БД всех заказов со статусом NEW или PROCESSING
// и последующей передаче их в канал
func (db *DBStor) GetAccurOrders(num int) chan models.OrderAns {
	numCh := make(chan models.OrderAns, num)
	go func() {
		for {
			// формируется querry запрос
			rows, err := db.DB.Query(QuerryGetAccurOrders)
			if err != nil {
				logger.Log.Error("Error in creating Query", zap.Error(err))
				rows.Close()
				return
			}
			for rows.Next() {
				for range num {
					db.Wg.Add(1)
					// получение данных из ряда
					var AnsOrd models.OrderAns
					err := rows.Scan(&AnsOrd.Number, &AnsOrd.StatusOld, &AnsOrd.UserID)
					if err != nil {
						logger.Log.Error("Error in Scan Query", zap.Error(err))
						break
					}
					// передача данных в канал для отправки к Accural
					numCh <- AnsOrd
					if !rows.Next() {
						break
					}
				}
				// ожидание флага о записи обновленных данных
				db.Wg.Wait()
			}
			if err := rows.Err(); err != nil {
				logger.Log.Error("Error in rows", zap.Error(err))
				rows.Close()
				continue
			}
			rows.Close()
		}
	}()
	return numCh
}

// функция для записи обновленных данных в базу в таблицы users и orders
func (db *DBStor) SetAccurOrders(writeCh chan models.OrderAns) {
	go func() {
		for toWrite := range writeCh {
			// проверяем обновился ли статус заказа
			if toWrite.Status != toWrite.StatusOld {
				// для нового статуса записываем новые данные в базу о заказах
				_, err := db.DB.Exec(QuerrySaveAccurOrders, toWrite.Status, toWrite.Accrual, toWrite.Number)
				if err != nil {
					logger.Log.Error("Error in rows", zap.Error(err))
				}
				// если статус операции успешеный, прибавляеми сумму баллов на баланс пользователя
				if toWrite.Status == "PROCESSED" {
					// функция для обнловления баланса пользователя с прибавлением полученных баллов
					err := db.AddUserBalance(toWrite.Accrual, toWrite.UserID)
					if err != nil {
						logger.Log.Error("Error in AddUserBalance", zap.Error(err))
					}
				}
			}
			// уведомляем стартовую горутину об успешной операции записи
			db.Wg.Done()
			time.Sleep(50 * time.Millisecond)
		}
	}()

}

// функция транзакционного списания средств с баланса пользователя
// и транзакционное сохранение нового заказа в базу с соответствующей меткой
func (db *DBStor) SaveWithdrawOrder(UserID string, number string, sum float64) error {
	var userBalance float64
	uploaded := time.Now()

	tx1, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx1.Rollback()
	tx2, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx2.Rollback()

	// вычитаем сумму у баланса пользователя в базе в первой транзакции
	row := tx1.QueryRow(QuerryMinusUserSum, sum, UserID)
	// записываем данные о заказе во второй транзакции
	_, err = tx2.Exec(QuerrySaveWithdrawOrder, UserID, number, sum, uploaded)
	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			return models.ErrorAlreadyTakenNumber
		} else {
			return err
		}
	}

	err = row.Scan(&userBalance)
	if err != nil {
		return err
	}
	// если баланс пользователя после списания стал отрицательным, откатываем транзакции
	if userBalance < 0 {
		return models.ErrorNoSuchBalance
	}
	err1 := tx1.Commit()
	if err1 != nil {
		return err1
	}
	err2 := tx2.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}

// функция для получения информации о всех заказах пользователя из базы
func (db *DBStor) GetUserWithdrawals(UserID string) (models.UserWithdrawList, error) {
	var err error
	var withdrawList models.UserWithdrawList
	var OpTime time.Time
	rows, err := db.DB.Query(QuerryGetUserWithdraw, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdraw models.UserWithdraw
		err := rows.Scan(&withdraw.Number, &withdraw.Accrual, &OpTime)
		if err != nil {
			return nil, err
		}
		withdraw.LoadedTime = OpTime.Format(time.RFC3339)
		withdrawList = append(withdrawList, withdraw)
	}
	if err := rows.Err(); err != nil { // (5)
		return nil, err
	}

	if len(withdrawList) == 0 {
		return nil, models.ErrorNoOrdersUser
	}
	return withdrawList, nil
}
