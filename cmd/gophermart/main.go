package main

import (
	"gophermart/internal/server"
)

func main() {
	err := server.RunServer()
	if err != nil {
		panic(err)
	}
}
