package models

import (
	"errors"

	"golang.org/x/exp/rand"
)

type UserRigisterLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserInfo struct {
	UserID  string
	Session bool
}

type UserOrder struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	LoadedTime string  `json:"uploaded_at"`
}

type UserOrdersList []UserOrder

type OrderAns struct {
	Number    string  `json:"order"`
	Status    string  `json:"status"`
	Accrual   float64 `json:"accrual"`
	StatusOld string  `json:"-"`
	UserID    string  `json:"-"`
}

type WithData struct {
	Number string  `json:"order"`
	Sum    float64 `json:"sum"`
}

type UserWithdraw struct {
	Number     string  `json:"order"`
	Accrual    float64 `json:"sum,omitempty"`
	LoadedTime string  `json:"uploaded_at"`
}

type UserWithdrawList []UserWithdraw

type BalanceData struct {
	Sum      float64 `json:"current"`
	Withdraw float64 `json:"withdrawn"`
}

func RandomString(n int) string {

	var bt []byte
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for range n {
		bt = append(bt, charset[rand.Intn(len(charset))])
	}

	return string(bt)
}

var ErrorNoUserString = "no rows in result set"
var ErrorBadToken = errors.New("token is not valid")
var ErrorUserAlreadyHas = errors.New("user already downloaded this number")
var ErrorAnotherUserHas = errors.New("another user downloaded this number")
var ErrorNoOrdersUser = errors.New("current user has no orders")
var ErrorNoSuchBalance = errors.New("not enough points on user balance")
var ErrorAlreadyTakenNumber = errors.New("someone already downloaded this number")

type KeyContext string

var CtxKey = KeyContext("UserID")
