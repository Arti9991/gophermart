package handlers

type HandlersData struct {
	BaseAddr  string
	AccurAddr string
	//DbStor
}

func HandlersDataInit(BaseAddr string, AccurAddr string) *HandlersData {
	return &HandlersData{BaseAddr: BaseAddr, AccurAddr: AccurAddr}
}
