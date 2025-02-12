package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophermart/internal/models"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequestRegistration(t *testing.T, ts *httptest.Server,
	path string, body io.Reader) *http.Response {
	client := http.DefaultClient

	request, err := http.NewRequest(http.MethodPost, ts.URL+path, body)
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(request)
	require.NoError(t, err)

	return resp
}

func testRequest(t *testing.T, ts *httptest.Server,
	method string, headers map[string]string,
	Cookie *http.Cookie,
	path string, body io.Reader) *http.Response {
	client := http.DefaultClient
	request, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	request.AddCookie(Cookie)

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	resp, err := client.Do(request)
	require.NoError(t, err)

	return resp
}

func TestRouter(t *testing.T) {
	serv := InitServer()
	ts := httptest.NewServer(serv.MainRouter())
	go AccrRun(&serv)
	defer ts.Close()

	type want struct {
		statusCodeOK       int
		statusCodeAccepted int
		statusCodePayment  int
	}
	tests := []struct {
		name                string
		requestRegistr      string
		requestLogin        string
		requestOrder        string
		requestCheckBalance string
		requestWutdraw      string
		UserInfo            string
		orders              []string
		UserOrderWithdraw   string
		want                want
	}{
		{
			name:                "Simple request for code 307",
			requestRegistr:      "/api/user/register",
			requestLogin:        "/api/user/login",
			requestOrder:        "/api/user/orders",
			requestCheckBalance: "/api/user/balance",
			requestWutdraw:      "/api/user/balance/withdraw",
			UserInfo:            `{"login":"UserTest","password":"12345678"}`,
			orders: []string{"7622483",
				"3421252",
				"57886311265",
				"137561315169",
				"856145766023",
				"1277617534",
				"138503636",
				"142434877838532",
				"710535703077780",
				"7682321",
				"6234728",
				"664101141102",
				"742854854",
				"6367437",
				"88130601670318",
			},
			UserOrderWithdraw: "605886811",
			want: want{
				statusCodeOK:       200,
				statusCodeAccepted: 202,
				statusCodePayment:  402,
			},
		},
	}
	for _, test := range tests {
		//ident := make([]string, 0)
		//locate := make([]string, 0)
		//for i := range len(test.bodys) {

		// запрос на регистрацию (если зарегистрирован, то уже на логин)
		resp := testRequestRegistration(t, ts, test.requestRegistr, bytes.NewBuffer([]byte(test.UserInfo)))
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusConflict {
			resp = testRequestRegistration(t, ts, test.requestLogin, bytes.NewBuffer([]byte(test.UserInfo)))
		}
		assert.Equal(t, test.want.statusCodeOK, resp.StatusCode)
		// проверяем наличие cookie
		assert.True(t, len(resp.Cookies()) > 0)
		tokenCookie := resp.Cookies()[0]
		// запрос на размещение заказов с куки предыдущего запроса
		for _, order := range test.orders {
			header := map[string]string{"Content-Type": "text/plain"}
			resp := testRequest(t, ts, http.MethodPost, header, tokenCookie, test.requestOrder, strings.NewReader(order))
			assert.True(t, (resp.StatusCode == test.want.statusCodeAccepted || resp.StatusCode == test.want.statusCodeOK))
		}
		time.Sleep(10000 * time.Microsecond)
		// запрос на получение баланса пользователя
		resp = testRequest(t, ts, http.MethodGet, nil, tokenCookie, test.requestCheckBalance, nil)
		assert.Equal(t, test.want.statusCodeOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var outBuff1 models.BalanceData
		json.NewDecoder(resp.Body).Decode(&outBuff1)
		fmt.Println(outBuff1)

		// запрос на списание средств больших чем баланс пользователя (должен не списать)
		var WithdrawUser models.WithData
		WithdrawUser.Number = test.UserOrderWithdraw
		WithdrawUser.Sum = outBuff1.Sum + 10
		reqBuff, err := json.Marshal(&WithdrawUser)
		require.NoError(t, err)
		header := map[string]string{"Content-Type": "application/json"}
		resp = testRequest(t, ts, http.MethodPost, header, tokenCookie, test.requestWutdraw, bytes.NewBuffer(reqBuff))
		assert.Equal(t, test.want.statusCodePayment, resp.StatusCode)

		// запрос на списание средств меньших чем баланс пользователя (должен списать и записать заказ на списание)
		WithdrawUser.Sum = outBuff1.Sum - 20
		reqBuff, err = json.Marshal(&WithdrawUser)
		require.NoError(t, err)
		resp = testRequest(t, ts, http.MethodPost, header, tokenCookie, test.requestWutdraw, bytes.NewBuffer(reqBuff))
		assert.Equal(t, test.want.statusCodeOK, resp.StatusCode)

		// запрос на проверку баланса пользователя
		resp = testRequest(t, ts, http.MethodGet, nil, tokenCookie, test.requestCheckBalance, nil)
		assert.Equal(t, test.want.statusCodeOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var outBuff2 models.BalanceData
		json.NewDecoder(resp.Body).Decode(&outBuff2)
		fmt.Println(outBuff2)
		// сравниваем баланс с разнцией в запроса на списание
		assert.Equal(t, 20.00, outBuff2.Sum)
	}
}
