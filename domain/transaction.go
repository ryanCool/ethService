package domain

import (
	"context"
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

type Receipt struct {
	TxHash string
}

type TransactionLog struct {
	TxHash   string `json:"-"`
	LogIndex int    `json:"index"`
	LogData  []byte `json:"data"`
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetTxHashesByBlockHash(ctx context.Context, blockHash string) ([]string, error)
	SaveReceiptAndLogs(ctx context.Context, txHash string, logs []TransactionLog) error
	GetLogsByTxHash(ctx context.Context, txHash string) ([]TransactionLog, error)
	GetByTxHash(ctx context.Context, txHash string) (*Transaction, error)
}

type TransactionUseCase interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetTxHashesByBlockHash(ctx context.Context, blockHash string) ([]string, error)
	SaveReceiptAndLogs(ctx context.Context, txHash string, logs []TransactionLog) error
	GetByTxHash(ctx context.Context, txHash string) (*Transaction, error)
}
