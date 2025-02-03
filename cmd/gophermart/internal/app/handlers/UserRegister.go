package handlers

import "net/http"

// хэндлер для получения оригинального URL по укороченному
func UserRegister(hd *HandlersData) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Test Answer UserRegister"))
	}

}
