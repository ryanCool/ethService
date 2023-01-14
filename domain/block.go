package domain

import "context"

type Block struct {
	Id         int    `json:"id"`
	BlockNum   int    `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  int    `json:"block_time"`
	ParentHash string `json:"parent_hash"`
}

type BlockRepository interface {
	Create(ctx context.Context, block *Block) error
}

type BlockUseCase interface {
	NewBlock(ctx context.Context, block *Block) error
}
