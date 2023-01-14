package usecase

import (
	"context"
	"github.com/ryanCool/ethService/domain"

	"time"
)

type blockUseCase struct {
	repo           domain.BlockRepository
	contextTimeout time.Duration
}

func NewBlockUseCase(a domain.BlockRepository, timeout time.Duration) domain.BlockUseCase {
	return &blockUseCase{
		repo:           a,
		contextTimeout: timeout,
	}
}

func (bu *blockUseCase) NewBlock(ctx context.Context, block *domain.Block) error {
	return bu.repo.Create(ctx, block)
}
