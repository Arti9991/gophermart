package storage

import "gophermart/internal/models"

type StorUserFunc interface {
	SaveUser(Login string, Password string, UserID string) error
	GetUserID(Login string) (string, string, error)
	MinusUserBalance(sum float64, UserID string) error
	GetUserBalance(UserID string) (models.BalanceData, error)
}

type StorOrderFunc interface {
	SaveNewOrder(UserID string, number string) error
	GetUserOrders(UserID string) (models.UserOrdersList, error)
	GetAccurOrders(numCh chan models.OrderAns)
	SetAccurOrders(toWrite models.OrderAns)
	SaveWithdrawOrder(UserID string, number string, sum float64) error
	GetUserWithdrawals(UserID string) (models.UserWithdrawList, error)
}
