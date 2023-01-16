package usecase

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
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

func (tu *transactionUseCase) saveReceipt(ctx context.Context, txHash common.Hash) error {
	receipt, err := tu.rpcClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Err(err).Msg("save receipt fail when get receipt through rpc client")
		return err
	}

	logs := []domain.TransactionLog{}
	for _, l := range receipt.Logs {
		tl := domain.TransactionLog{
			TxHash:   txHash.String(),
			LogIndex: int(l.Index),
			LogData:  l.Data,
		}
		logs = append(logs, tl)
	}

	if len(logs) > 0 {
		err = tu.repo.SaveReceiptAndLogs(ctx, txHash.String(), logs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tu *transactionUseCase) Save(ctx context.Context, blockHash string, transaction *types.Transaction) error {
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

	go func() {
		err = tu.saveReceipt(ctx, transaction.Hash())
		if err != nil {
			log.Err(err).Msg("save receipt fail")
		}
	}()

	return nil
}
