package database

import (
	"database/sql"
	"gophermart/internal/logger"
	"gophermart/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var QuerryCreateUserStor = `CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
	user_id VARCHAR(16),
    login 	VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(64) NOT NULL,
	sum NUMERIC(10,2) NOT NULL,
	withdraw NUMERIC(10,2) NOT NULL
	);`
var QuerrySaveUser = `INSERT INTO users (id, user_id, login, password, sum, withdraw)
  	VALUES  (DEFAULT, $1, $2, $3, 0, 0);`
var QuerryGetUser = `SELECT user_id, password
	FROM users WHERE login = $1 LIMIT 1;`
var QuerryUpdateUserSum = `UPDATE users SET sum = sum + $1
	WHERE user_id = $2;`

type DBStor struct {
	storage.StorUserFunc
	storage.StorOrderFunc
	DB     *sql.DB
	DBInfo string
	flagCh chan struct{}
}

// инициализация хранилища и создание/подключение к таблице
func DBUserInit(DBInfo string) (*DBStor, error) {
	var db DBStor
	var err error

	db.DBInfo = DBInfo

	db.DB, err = sql.Open("pgx", DBInfo)
	if err != nil {
		return &DBStor{}, err
	}

	if err = db.DB.Ping(); err != nil {
		return &DBStor{}, err
	}

	_, err = db.DB.Exec(QuerryCreateUserStor)
	if err != nil {
		return &DBStor{}, err
	}
	db.flagCh = make(chan struct{})
	// go func() {
	// 	db.flagCh <- struct{}{}
	// }()
	logger.Log.Info("✓ created users table!")
	return &db, nil
}

func (db *DBStor) SaveUser(Login string, Password string, UserID string) error {

	var err error
	_, err = db.DB.Exec(QuerrySaveUser, UserID, Login, Password)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBStor) GetUserID(Login string) (string, string, error) {

	var err error
	var UserID string
	var Password string

	row := db.DB.QueryRow(QuerryGetUser, Login)
	err = row.Scan(&UserID, &Password)
	if err != nil {
		return "", "", err
	}

	return UserID, Password, nil
}

func (db *DBStor) AddUserBalance(sum float64, UserID string) error {

	_, err := db.DB.Exec(QuerryUpdateUserSum, sum, UserID)
	if err != nil {
		return err
	}

	return nil
}
