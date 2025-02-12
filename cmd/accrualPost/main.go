package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/exp/rand"
)

type goods struct {
	Name  string  `json:"description"`
	Price float64 `json:"price"`
}

type Order struct {
	Number string `json:"order"`
	Goods  []goods
}

type GoodType struct {
	Name    string `json:"match"`
	Reward  int    `json:"reward"`
	RewType string `json:"reward_type"`
}

func main() {
	types, err := NewGoodsTypes("http://localhost:8083/")
	if err != nil {
		panic(err)
	}
	err = PostNumberToAPI("http://localhost:8083/", 25, types)
	if err != nil {
		panic(err)
	}
}

// заполняем accural товарами и суммой их вознагпаждения
func NewGoodsTypes(AccAddr string) ([]GoodType, error) {
	var MassGoods []GoodType

	var good1 GoodType
	good1.Name = "Bosh"
	good1.Reward = 30
	good1.RewType = "%"
	MassGoods = append(MassGoods, good1)
	var good2 GoodType
	good2.Name = "Braun"
	good2.Reward = 10
	good2.RewType = "%"
	MassGoods = append(MassGoods, good2)
	var good3 GoodType
	good3.Name = "Electrolux"
	good3.Reward = 100
	good3.RewType = "pt"
	MassGoods = append(MassGoods, good3)
	var good4 GoodType
	good4.Name = "Apple"
	good4.Reward = 300
	good4.RewType = "pt"
	MassGoods = append(MassGoods, good4)
	var good5 GoodType
	good5.Name = "Samsung"
	good5.Reward = 50
	good5.RewType = "pt"
	MassGoods = append(MassGoods, good5)
	var good6 GoodType
	good6.Name = "Xioami"
	good6.Reward = 10
	good6.RewType = "%"
	MassGoods = append(MassGoods, good6)
	var good7 GoodType
	good7.Name = "TV"
	good7.Reward = 25
	good7.RewType = "%"
	MassGoods = append(MassGoods, good7)

	for _, good := range MassGoods {
		buff, err := json.Marshal(good)
		if err != nil {
			return nil, err
		}
		client := http.Client{
			Timeout: time.Second * 3, // интервал ожидания: 1 секунда
		}
		reqURI := AccAddr + "api/goods"
		request, err := http.NewRequest(http.MethodPost, reqURI, bytes.NewReader(buff))
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		// выводим код ответа
		fmt.Println("Статус-код ", response.Status)
		defer response.Body.Close()
	}
	return MassGoods, nil
}

// создаем список возможных заказов, которые доступны в accrual
func PostNumberToAPI(AccAddr string, n int, Types []GoodType) error {
	file, err := os.OpenFile("orders_numbers.csv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Print("ERROR IN FILE OPEN")
	}
	writer := bufio.NewWriter(file)
	//var outBuff OrderAns
	for range n {
		num := GenerateOrderNumber()

		var Order Order
		Order.Number = num
		gNum := (rand.Intn(7) + 1)
		for range gNum {
			var good goods
			good.Name = Types[rand.Intn(6)].Name
			good.Price = float64((rand.Intn(100000) + 1000))
			Order.Goods = append(Order.Goods, good)
		}
		fmt.Println(Order)

		buff, err := json.Marshal(Order)
		if err != nil {
			return err
		}

		client := http.Client{
			Timeout: time.Second * 3, // интервал ожидания: 3 секунды
		}

		reqURI := AccAddr + "api/orders"
		fmt.Println(reqURI)
		request, err := http.NewRequest(http.MethodPost, reqURI, bytes.NewReader(buff))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			return err
		}

		defer response.Body.Close()
		if response.StatusCode != 202 {
			fmt.Println(response.Status)
			return errors.New("bad responce status")
		}

		wr := []byte(num + "\n")
		_, err = writer.Write(wr)
		if err != nil {
			fmt.Print("ERROR IN FILE WRITE")
		}

	}
	err = writer.Flush()
	if err != nil {
		fmt.Print("ERROR IN FILE FLUSH")
	}
	return nil
}
