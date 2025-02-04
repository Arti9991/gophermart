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

type KeyContext string

var CtxKey = KeyContext("UserID")
