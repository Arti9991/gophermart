package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"gophermart/internal/storage"
)

type HandlersData struct {
	BaseAddr  string
	AccurAddr string
	StorFunc  storage.StorFunc
	//DbStor
}

func HandlersDataInit(BaseAddr string, AccurAddr string, Stor storage.StorFunc) *HandlersData {
	return &HandlersData{BaseAddr: BaseAddr, AccurAddr: AccurAddr, StorFunc: Stor}
}

// функция хэширования для паролей (sha256)
func CodePassword(src string) string {
	hash := sha256.New()
	hash.Write([]byte(src))
	dst := hash.Sum(nil)
	return hex.EncodeToString(dst)
}
