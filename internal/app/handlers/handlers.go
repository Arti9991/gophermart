package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"gophermart/internal/storage"
)

type HandlersData struct {
	BaseAddr  string
	AccurAddr string
	StorUser  storage.StorUserFunc
	StorOrder storage.StorOrderFunc
	//DbStor
}

func HandlersDataInit(BaseAddr string, AccurAddr string, StorUser storage.StorUserFunc, StorOrder storage.StorOrderFunc) *HandlersData {
	return &HandlersData{BaseAddr: BaseAddr, AccurAddr: AccurAddr, StorUser: StorUser, StorOrder: StorOrder}
}

// функция хэширования для паролей (sha256)
func CodePassword(src string) string {
	hash := sha256.New()
	hash.Write([]byte(src))
	dst := hash.Sum(nil)
	return hex.EncodeToString(dst)
}
