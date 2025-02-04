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
    password VARCHAR(64) NOT NULL
	);`
var QuerrySaveUser = `INSERT INTO users (id, user_id, login, password)
  	VALUES  (DEFAULT, $1, $2, $3);`
var QuerryGetUser = `SELECT user_id, password
	FROM users WHERE login = $1 LIMIT 1;`

type DBStor struct {
	storage.StorFunc
	DB     *sql.DB
	DBInfo string
}

// инициализация хранилища и создание/подключение к таблице
func DBinit(DBInfo string) (*DBStor, error) {
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
	logger.Log.Info("✓ connected to ShortURL db!")
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
