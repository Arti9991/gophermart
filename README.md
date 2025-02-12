# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.

# Тестовые запросы для ручной проверки сервера

Запрос на регистрацию нового пользователя:

```
curl -v -X POST -H "Content-Type: application/json" -d '{"login":"User","password":"12345678"}' http://localhost:8082/api/user/register
```
Запрос на вход пользователя:
```
curl -v -X POST -H "Content-Type: application/json" -d '{"login":"User","password":"12345678"}' http://localhost:8082/api/user/login
```

Запрос на отправку номера заказа от пользователя
```
curl -v -X POST -H "Content-Type: text/plain" --cookie "Token=<cookie>" -d 1234567890101112131415 http://localhost:8082/api/user/orders
```
Запрос на на получение всех заказов пользователя
```
curl -v -X GET --cookie "Token=<cookie>" http://localhost:8082/api/user/orders
```

Запрос на списание средств с накопительного счета пользователя
```
curl -v -X POST -H "Content-Type: application/json" --cookie "Token=<cookie>" -d '{"order":<order_num>,"sum": 500}' http://localhost:8082/api/user/balance/withdraw
```
Запрос на отображение баланса пользователя
```
curl -v -X GET --cookie "Token=<cookie>" http://localhost:8082/api/user/balance
```
Запрос на отображение всех заказов со списанием у пользователя
```
curl -v -X GET --cookie "Token=<cookie>" http://localhost:8082/api/user/withdrawals
```

./gophermarttest.exe --test.v --test.run=^TestGophermart$ --gophermart-binary-path=cmd/gophermart/gophermart.exe --gophermart-host=localhost --gophermart-port=8082 --gophermart-database-uri="host=localhost user=myuser password=123456 dbname=Gophermart sslmode=disable" --accrual-binary-path=cmd/accrual/accrual_windows_amd64.exe -accrual-host=localhost -accrual-port=8083 -accrual-database-uri="host=localhost user=myuser password=123456 dbname=Gophermart sslmode=disable"

