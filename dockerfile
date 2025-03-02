FROM golang:1.22

WORKDIR /usr/src/gophermart

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /usr/src/gophermart/cmd/gophermart
EXPOSE 8082
ENV RUN_ADDRESS=:8082

RUN go build -v -o /usr/local/bin/gophermart ./...

ENTRYPOINT ["gophermart"]