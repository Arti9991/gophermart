package database

import (
	"gophermart/internal/logger"
	"gophermart/internal/models"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var QuerryCreateTypeStatus = `CREATE TYPE status AS ENUM('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');`
var QuerryCreateorderStor = `
	CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
	user_id VARCHAR(16) NOT NULL,
	number VARCHAR(32) NOT NULL UNIQUE,
    status status NOT NULL,
    accrual DECIMAL(10,2) NOT NULL,
	uploaded_at TIMESTAMPTZ
	);`

var QuerrySaveNewOrder = `INSERT INTO orders (id, user_id, number, status, accrual, uploaded_at)
  	VALUES  (DEFAULT, $1, $2, 'NEW', 0.0, $3);`
var QuerryGetUserForNumber = `SELECT user_id FROM orders 
	WHERE number = $1 LIMIT 1;`
var QuerryGetUserOrders = `SELECT number, status, accrual, uploaded_at FROM orders 
	WHERE user_id = $1;`

// var QuerryGetOrder = `SELECT user_id, password
// 	FROM users WHERE login = $1 LIMIT 1;`

// var QuerryUpdateUserSum = `UPDATE users SET sum = ($1)
// 	WHERE user_id = ($2);`

// инициализация хранилища и создание/подключение к таблице
func (db *DBStor) DBOrdersInit() error {

	_, err := db.DB.Exec(QuerryCreateorderStor)
	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UndefinedObject) {
			_, err = db.DB.Exec(QuerryCreateTypeStatus)
			if err != nil {
				return err
			}
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

func (db *DBStor) SaveNewOrder(UserID string, number string) error {
	uploaded := time.Now()
	var err error
	var UserID2 string
	_, err = db.DB.Exec(QuerrySaveNewOrder, UserID, number, uploaded)
	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			row := db.DB.QueryRow(QuerryGetUserForNumber, number)
			err = row.Scan(&UserID2)
			if err != nil {
				return err
			}
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

func (db *DBStor) GetUserOrders(UserID string) (models.UserOrdersList, error) {
	var err error
	var ordersList models.UserOrdersList
	var OpTime time.Time
	rows, err := db.DB.Query(QuerryGetUserOrders, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.UserOrder
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &OpTime)
		if err != nil {
			return nil, err
		}
		order.LoadedTime = OpTime.Format(time.RFC3339)
		ordersList = append(ordersList, order)
	}
	if err := rows.Err(); err != nil { // (5)
		return nil, err
	}

	if len(ordersList) == 0 {
		return nil, models.ErrorNoOrdersUser
	}
	return ordersList, nil
}

// func (db *DBStor) GetUserID(Login string) (string, string, error) {

// 	var err error
// 	var UserID string
// 	var Password string

// 	row := db.DB.QueryRow(QuerryGetUser, Login)
// 	err = row.Scan(&UserID, &Password)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	return UserID, Password, nil
// }
