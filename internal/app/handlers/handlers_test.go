package handlers

import (
	"bytes"
	"errors"
	"gophermart/internal/models"
	"gophermart/internal/storage/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
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
