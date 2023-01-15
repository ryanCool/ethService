package domain

type Transaction struct {
	TxHash string           `json:"tx_hash"`
	From   string           `json:"from"`
	To     string           `json:"to"`
	Nonce  uint64           `json:"nonce"`
	Data   string           `json:"data"`
	Value  string           `json:"value" json:"value"`
	Logs   []TransactionLog `json:"logs" json:"logs"`
}

type TransactionLog struct {
	Index int
	Data  string
}
