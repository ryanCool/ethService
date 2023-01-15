package domain

import (
	"context"
)

type Block struct {
	BlockDb
	TransactionHashes []string `json:"transactions" gorm:"-"`
}

type BlockDb struct {
	BlockNum   uint64 `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
	Stable     bool   `json:"stable"`
}

type BlockRepository interface {
	List(ctx context.Context, limit int) ([]BlockDb, error)
	Create(ctx context.Context, block *BlockDb) error
	SetStable(ctx context.Context, blockNum uint64, stable bool) error
	GetByNumber(ctx context.Context, blockNum uint64) (*BlockDb, error)
}

type BlockUseCase interface {
	List(ctx context.Context, limit int) ([]BlockDb, error)
	GetByNumber(ctx context.Context, blockNum uint64) (*Block, error)
	Initialize(ctx context.Context)
}
