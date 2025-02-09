package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gophermart/internal/models"
	"gophermart/internal/storage/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//для базовых тестов производится генерация моков командой ниже
// mockgen --source=./internal/storage/storage.go --destination=./internal/storage/mocks/mocks_store.go --package=mocks StorFunc StorOrder

var BaseAdr = "http://example.com"

var UserID = "125"

func TestRegistration(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		err        error
	}

	tests := []struct {
		name    string
		request string
		body    string
		want    want
	}{
		{
			name:    "Simple request for registration and code 200",
			request: "/api/user/register",
			body:    `{"login":"User","password":"12345678"}`,
			want: want{
				statusCode: 200,
				err:        nil,
			},
		},
		{
			name:    "Request with big password and login for registration and code 200",
			request: "/api/user/register",
			body:    `{"login":"UserDontWantToRegistrJustPlayNotFunnyGames","password":"123456789101112131231421414213"}`,
			want: want{
				statusCode: 200,
				err:        nil,
			},
		},
		{
			name:    "Request with DB error",
			request: "/api/user/register",
			body:    `{}`,
			want: want{
				statusCode: 500,
				err:        errors.New(pgerrcode.QueryCanceled),
			},
		},
		{
			name:    "Request with already taken login",
			request: "/api/user/register",
			body:    `{"login":"User","password":"12345678"}`,
			want: want{
				statusCode: 409,
				err:        errors.New(pgerrcode.UniqueViolation),
			},
		},
		{
			name:    "Request with bad body",
			request: "/api/user/register",
			body:    `{12345}`,
			want: want{
				statusCode: 400,
				err:        nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m1.EXPECT().
				SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodPost, test.request, bytes.NewBuffer([]byte(test.body)))
			request.Header.Add("Content-Type", "application/json")
			// ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: UserID})
			// request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(UserRegister(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

		})
	}
}

func TestLogin(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		userID     string
		password   string
		err        error
	}

	tests := []struct {
		name    string
		request string
		body    string
		want    want
	}{
		{
			name:    "Simple request for login and code 200",
			request: "/api/user/login",
			body:    `{"login":"User","password":"userPassword"}`,
			want: want{
				statusCode: 200,
				userID:     "userID",
				password:   "userPassword",
				err:        nil,
			},
		},
		{
			name:    "Request with DB error",
			request: "/api/user/login",
			body:    `{"login":"User","password":"userPassword"}`,
			want: want{
				statusCode: 500,
				userID:     "",
				password:   "",
				err:        errors.New(pgerrcode.UniqueViolation),
			},
		},
		{
			name:    "Request with big password and login for registration and code 200",
			request: "/api/user/login",
			body:    `{"login":"UserDontWantToRegistrJustPlayNotFunnyGames","password":"123456789101112131231421414213"}`,
			want: want{
				statusCode: 200,
				userID:     "userID",
				password:   "123456789101112131231421414213",
				err:        nil,
			},
		},
		{
			name:    "Request with bad login",
			request: "/api/user/login",
			body:    `{"login":"UserBad","password":"12345678"}`,
			want: want{
				statusCode: 401,
				userID:     "",
				password:   "",
				err:        errors.New(models.ErrorNoUserString),
			},
		},
		{
			name:    "Request with bad body",
			request: "/api/user/login",
			body:    `{12345}`,
			want: want{
				statusCode: 400,
				err:        nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// кодирование пароля из теста
			test.want.password = CodePassword(test.want.password)
			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m1.EXPECT().
				GetUserID(gomock.Any()).
				Return(test.want.userID, test.want.password, test.want.err).
				MaxTimes(1)
			request := httptest.NewRequest(http.MethodPost, test.request, bytes.NewBuffer([]byte(test.body)))
			request.Header.Add("Content-Type", "application/json")
			// ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: UserID})
			// request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(UserLogin(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

		})
	}
}

func TestPostOrder(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		err        error
	}

	tests := []struct {
		name    string
		userID  string
		request string
		body    string
		want    want
	}{
		{
			name:    "Request for code 500",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "12345678903",
			want: want{
				statusCode: 500,
				err:        errors.New("some sql error"),
			},
		},
		{
			name:    "Simple request for code 202",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "12345678903",
			want: want{
				statusCode: 202,
				err:        nil,
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "12345678903",
			want: want{
				statusCode: 200,
				err:        models.ErrorUserAlreadyHas,
			},
		},
		{
			name:    "Request for code 409",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "12345678903",
			want: want{
				statusCode: 409,
				err:        models.ErrorAnotherUserHas,
			},
		},
		{
			name:    "Request for code 422",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "123456",
			want: want{
				statusCode: 422,
				err:        models.ErrorAnotherUserHas,
			},
		},
		{
			name:    "Request for code 400",
			userID:  "userID",
			request: "/api/user/orders",
			body:    "",
			want: want{
				statusCode: 400,
				err:        nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m2.EXPECT().
				SaveNewOrder(gomock.Any(), gomock.Any()).
				Return(test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodPost, test.request, strings.NewReader(test.body))

			ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: test.userID})
			request = request.WithContext(ctx)

			request.Header.Add("Content-Type", "text/plain")
			// ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: UserID})
			// request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostOrder(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

		})
	}
}

func TestGetOrder(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		answer     models.UserOrdersList
		err        error
	}

	tests := []struct {
		name    string
		userID  string
		request string
		want    want
	}{
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/orders",
			want: want{
				statusCode: 200,
				answer: []models.UserOrder{
					{Number: "12345678903", Status: "NEW", Accrual: 0.0, LoadedTime: time.Now().Format(time.RFC3339)},
				},
				err: nil,
			},
		},
		{
			name:    "Request for code 200",
			userID:  "userID",
			request: "/api/user/orders",
			want: want{
				statusCode: 200,
				answer: []models.UserOrder{
					{Number: "12345678903", Status: "NEW", Accrual: 0.0, LoadedTime: time.Now().Format(time.RFC3339)},
					{Number: "346436439", Status: "PROCESSED", Accrual: 5000.0, LoadedTime: time.Now().Format(time.RFC3339)},
					{Number: "9278923470", Status: "INVALID", Accrual: 100.5, LoadedTime: time.Now().Format(time.RFC3339)},
				},
				err: nil,
			},
		},
		{
			name:    "Request for code 204",
			userID:  "userID",
			request: "/api/user/orders",
			want: want{
				statusCode: 204,
				answer:     nil,
				err:        models.ErrorNoOrdersUser,
			},
		},
		{
			name:    "Request for code 500",
			userID:  "userID",
			request: "/api/user/orders",
			want: want{
				statusCode: 500,
				answer:     nil,
				err:        errors.New("some sql error"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m2.EXPECT().
				GetUserOrders(gomock.Any()).
				Return(test.want.answer, test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: test.userID})
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetOrders(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

			var outBuff models.UserOrdersList
			json.NewDecoder(result.Body).Decode(&outBuff)

			assert.Equal(t, test.want.answer, outBuff)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestUserWithdraw(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		err        error
	}

	tests := []struct {
		name    string
		userID  string
		request string
		body    string
		want    want
	}{
		{
			name:    "Simple request for code 500",
			userID:  "userID",
			request: "/api/user/balance/withdraw",
			body:    `{"order":"137561315169","sum":2500}`,
			want: want{
				statusCode: 500,
				err:        errors.New("some sql error"),
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/balance/withdraw",
			body:    `{"order":"137561315169","sum":250}`,
			want: want{
				statusCode: 200,
				err:        nil,
			},
		},
		{
			name:    "Simple request for code 402",
			userID:  "userID",
			request: "/api/user/balance/withdraw",
			body:    `{"order":"137561315169","sum":2500}`,
			want: want{
				statusCode: 402,
				err:        models.ErrorNoSuchBalance,
			},
		},
		{
			name:    "Simple request for code 422 already taken",
			userID:  "userID",
			request: "/api/user/balance/withdraw",
			body:    `{"order":"137561315169","sum":2500}`,
			want: want{
				statusCode: 422,
				err:        models.ErrorAlreadyTakenNumber,
			},
		},
		{
			name:    "Simple request for code 422 luhn",
			userID:  "userID",
			request: "/api/user/balance/withdraw",
			body:    `{"order":"12345","sum":2500}`,
			want: want{
				statusCode: 422,
				err:        nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m2.EXPECT().
				SaveWithdrawOrder(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodPost, test.request, bytes.NewBuffer([]byte(test.body)))
			request.Header.Add("Content-Type", "application/json")

			ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: test.userID})
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(WithdrawOrder(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestGetWithrawals(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		answer     models.UserWithdrawList
		err        error
	}

	tests := []struct {
		name    string
		userID  string
		request string
		want    want
	}{
		{
			name:    "Request for code 500",
			userID:  "userID",
			request: "/api/user/withdrawals",
			want: want{
				statusCode: 500,
				answer:     nil,
				err:        errors.New("some sql error"),
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/withdrawals",
			want: want{
				statusCode: 200,
				answer: []models.UserWithdraw{
					{Number: "12345678903", Accrual: 0.0, LoadedTime: time.Now().Format(time.RFC3339)},
				},
				err: nil,
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/withdrawals",
			want: want{
				statusCode: 200,
				answer: []models.UserWithdraw{
					{Number: "12345678903", Accrual: 100.0, LoadedTime: time.Now().Format(time.RFC3339)},
					{Number: "346436439", Accrual: 500.0, LoadedTime: time.Now().Format(time.RFC3339)},
					{Number: "9278923470", Accrual: 10000.0, LoadedTime: time.Now().Format(time.RFC3339)},
				},
				err: nil,
			},
		},
		{
			name:    "Simple request for code 204",
			userID:  "userID",
			request: "/api/user/withdrawals",
			want: want{
				statusCode: 204,
				answer:     nil,
				err:        models.ErrorNoOrdersUser,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m2.EXPECT().
				GetUserWithdrawals(gomock.Any()).
				Return(test.want.answer, test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: test.userID})
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetWithrawals(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

			var outBuff models.UserWithdrawList
			json.NewDecoder(result.Body).Decode(&outBuff)

			assert.Equal(t, test.want.answer, outBuff)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestGetBalance(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m1 := mocks.NewMockStorUserFunc(ctrl)
	m2 := mocks.NewMockStorOrderFunc(ctrl)
	// инциализация handlers data для тестов
	hd := HandlersDataInit(BaseAdr, "", m1, m2)

	type want struct {
		statusCode int
		answer     models.BalanceData
		err        error
	}

	tests := []struct {
		name    string
		userID  string
		request string
		want    want
	}{
		{
			name:    "Request for code 500",
			userID:  "userID",
			request: "/api/user/balance",
			want: want{
				statusCode: 500,
				answer:     models.BalanceData{},
				err:        errors.New("some sql error"),
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/balance",
			want: want{
				statusCode: 200,
				answer:     models.BalanceData{Sum: 1500.10, Withdraw: 1000},
				err:        nil,
			},
		},
		{
			name:    "Simple request for code 200",
			userID:  "userID",
			request: "/api/user/balance",
			want: want{
				statusCode: 200,
				answer:     models.BalanceData{Sum: 150000.10, Withdraw: 10.05},
				err:        nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// задаем режим рабоыт моков (для POST главное отсутствие ошибки)
			m1.EXPECT().
				GetUserBalance(gomock.Any()).
				Return(test.want.answer, test.want.err).
				MaxTimes(1)

			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			ctx := context.WithValue(request.Context(), models.CtxKey, models.UserInfo{UserID: test.userID})
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetBalance(hd))
			h(w, request)

			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)

			var outBuff models.BalanceData
			json.NewDecoder(result.Body).Decode(&outBuff)

			assert.Equal(t, test.want.answer, outBuff)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
