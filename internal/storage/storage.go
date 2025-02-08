package storage

import "gophermart/internal/models"

type StorUserFunc interface {
	SaveUser(Login string, Password string, UserID string) error
	GetUserID(Login string) (string, string, error)
}

type StorOrderFunc interface {
	SaveNewOrder(UserID string, number string) error
	GetUserOrders(UserID string) (models.UserOrdersList, error)
	GetAccurOrders() chan models.OrderAns
	SetAccurOrders(inp chan models.OrderAns)
}
