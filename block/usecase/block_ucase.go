package usecase

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/ryanCool/ethService/domain"
	"gorm.io/gorm"
	"time"
)

type blockUseCase struct {
	repo             domain.BlockRepository
	transactionUcase domain.TransactionUseCase
	contextTimeout   time.Duration
}

func NewBlockUseCase(a domain.BlockRepository, t domain.TransactionUseCase, timeout time.Duration) domain.BlockUseCase {
	return &blockUseCase{
		repo:             a,
		transactionUcase: t,
		contextTimeout:   timeout,
	}
}

func (bu *blockUseCase) Create(ctx context.Context, block *domain.BlockDb) error {
	return bu.repo.Create(ctx, block)
}

func (bu *blockUseCase) DeleteByNum(ctx context.Context, blockNum uint64) error {
	return bu.repo.DeleteByNum(ctx, blockNum)
}

func (bu *blockUseCase) SetStable(ctx context.Context, blockNum uint64, stable bool) error {
	return bu.repo.SetStable(ctx, blockNum, stable)
}

//List list latest limit block
func (bu *blockUseCase) List(ctx context.Context, limit int) ([]domain.BlockDb, error) {
	return bu.repo.List(ctx, limit)
}

func (bu *blockUseCase) GetByNumber(ctx context.Context, blockNum uint64) (*domain.Block, error) {
	block, err := bu.repo.GetByNumber(ctx, blockNum)
	if err == gorm.ErrRecordNotFound {
		return nil, domain.ErrBlockNotExist
	}

	if err != nil {
		log.Err(err).Msg("get block by block_num fail")
		return nil, err
	}

	txs, err := bu.transactionUcase.GetTxHashesByBlockHash(ctx, block.BlockHash)
	if err != nil {
		log.Err(err).Msg("get tx hashes by block_hash fail")
		return nil, err
	}

	return &domain.Block{
		BlockDb:           *block,
		TransactionHashes: txs,
	}, nil
}
