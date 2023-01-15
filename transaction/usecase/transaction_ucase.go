package usecase

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ryanCool/ethService/domain"
	"time"
)

type transactionUseCase struct {
	repo             domain.TransactionRepository
	rpcClient        *ethclient.Client
	transactionUcase domain.TransactionUseCase
	contextTimeout   time.Duration
}

func NewTransactionUseCase(a domain.TransactionRepository, timeout time.Duration, rpcClient *ethclient.Client) domain.TransactionUseCase {
	return &transactionUseCase{
		repo:           a,
		contextTimeout: timeout,
		rpcClient:      rpcClient,
	}
}

func (tu *transactionUseCase) SaveTransaction(ctx context.Context, blockHash string, transaction *types.Transaction) error {
	from, err := types.Sender(types.LatestSignerForChainID(transaction.ChainId()), transaction)
	if err != nil {
		return err
	}

	to := transaction.To()
	if to == nil {
		to = &common.Address{}
	}
	err = tu.repo.Create(ctx, &domain.Transaction{
		BlockHash: blockHash,
		TxHash:    transaction.Hash().String(),
		TxFrom:    from.String(),
		TxTo:      to.String(),
		Nonce:     transaction.Nonce(),
		TxData:    transaction.Data(),
		TxValue:   transaction.Value().String(),
	})
	if err != nil {
		return err
	}

	return nil
}
