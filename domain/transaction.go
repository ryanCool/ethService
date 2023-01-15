package domain

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	BlockHash string           `json:"-"`
	TxHash    string           `json:"tx_hash"`
	TxFrom    string           `json:"from"`
	TxTo      string           `json:"to"`
	Nonce     uint64           `json:"nonce"`
	TxData    []byte           `json:"data"`
	TxValue   string           `json:"value"`
	Logs      []TransactionLog `json:"logs" gorm:"-"`
}

type TransactionLog struct {
	Index int
	Data  string
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
}

type TransactionUseCase interface {
	SaveTransaction(ctx context.Context, blockHash string, transaction *types.Transaction) error
}
