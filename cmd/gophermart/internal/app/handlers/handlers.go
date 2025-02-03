package handlers

import "gophermart/internal/storage"

type HandlersData struct {
	BaseAddr  string
	AccurAddr string
	StorFunc  storage.StorFunc
	//DbStor
}

func HandlersDataInit(BaseAddr string, AccurAddr string, Stor storage.StorFunc) *HandlersData {
	return &HandlersData{BaseAddr: BaseAddr, AccurAddr: AccurAddr, StorFunc: Stor}
}
