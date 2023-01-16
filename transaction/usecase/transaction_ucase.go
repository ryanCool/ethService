package usecase

import (
	"context"
	"github.com/ryanCool/ethService/domain"
	"time"
)

type transactionUseCase struct {
	repo             domain.TransactionRepository
	transactionUcase domain.TransactionUseCase
	contextTimeout   time.Duration
}

func NewTransactionUseCase(a domain.TransactionRepository, timeout time.Duration) domain.TransactionUseCase {
	return &transactionUseCase{
		repo:           a,
		contextTimeout: timeout,
	}
}

func (tu *transactionUseCase) Create(ctx context.Context, transaction *domain.Transaction) error {
	return tu.repo.Create(ctx, transaction)
}

func (tu *transactionUseCase) GetByTxHash(ctx context.Context, txHash string) (*domain.Transaction, error) {
	l, err := tu.repo.GetByTxHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	logs, err := tu.repo.GetLogsByTxHash(ctx, txHash)
	if err != nil {
		return nil, err
	}
	l.Logs = logs

	return l, nil
}

func (tu *transactionUseCase) GetTxHashesByBlockHash(ctx context.Context, blockHash string) ([]string, error) {
	return tu.repo.GetTxHashesByBlockHash(ctx, blockHash)
}

func (tu *transactionUseCase) SaveReceiptAndLogs(ctx context.Context, txHash string, logs []domain.TransactionLog) error {
	return tu.repo.SaveReceiptAndLogs(ctx, txHash, logs)
}
