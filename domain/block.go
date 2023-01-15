package domain

import (
	"context"
)

type Block struct {
	BlockDb
	Transactions []string `json:"transactions" gorm:"-"`
}

type BlockDb struct {
	BlockNum   uint64 `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
	Stable     bool   `json:"-"`
}

type BlockRepository interface {
	List(ctx context.Context, limit int) ([]BlockDb, error)
	Create(ctx context.Context, block *BlockDb) error
}

type BlockUseCase interface {
	List(ctx context.Context, limit int) ([]BlockDb, error)
	Initialize(ctx context.Context)
}
