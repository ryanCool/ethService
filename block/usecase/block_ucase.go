package usecase

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/domain"
	"math/big"
	"time"
)

type blockUseCase struct {
	repo             domain.BlockRepository
	rpcClient        *ethclient.Client
	wsClient         *ethclient.Client
	transactionUcase domain.TransactionUseCase
	contextTimeout   time.Duration
}

var syncFromNBlock *big.Int
var confirmedNum, scanWorkerNum int

func init() {
	confirmedNum = config.GetInt("CONFIRMED_BLOCK_NUM")
	scanWorkerNum = config.GetInt("SCAN_WORK_NUM")
	syncFromNBlock = config.GetBigInt("SYNC_BLOCK_FROM_N")
}

//Initialize init cron job to subscribe new block event through websocket endpoint
func (bu *blockUseCase) Initialize(ctx context.Context) {
	go bu.subscribeNewBlock(ctx)
	go bu.scanToLatest(ctx)
}

func NewBlockUseCase(a domain.BlockRepository, t domain.TransactionUseCase, timeout time.Duration, rpcClient *ethclient.Client, wsClient *ethclient.Client) domain.BlockUseCase {
	return &blockUseCase{
		repo:             a,
		transactionUcase: t,
		contextTimeout:   timeout,
		wsClient:         wsClient,
		rpcClient:        rpcClient,
	}
}

//NewBlock store block to db
func (bu *blockUseCase) newBlock(ctx context.Context, block *domain.BlockDb) error {
	return bu.repo.Create(ctx, block)
}

//setBlockStable set old block to stable status
func (bu *blockUseCase) setBlockStable(ctx context.Context, blockNum uint64, stable bool) error {
	return bu.repo.SetStable(ctx, blockNum, stable)
}

//List list latest limit block
func (bu *blockUseCase) List(ctx context.Context, limit int) ([]domain.BlockDb, error) {
	return bu.repo.List(ctx, limit)
}

//ScanToLatest scan blocks from n to latest , and store to db
func (bu *blockUseCase) scanToLatest(ctx context.Context) {
	header, err := bu.rpcClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error().Err(err)
	}
	latestNum := header.Number.Uint64()
	targetBlockNum := syncFromNBlock

	//use buffer channel to implement a worker pool with config number
	c := make(chan bool, scanWorkerNum)
	for targetBlockNum.Uint64() <= latestNum {
		c <- true
		go func(latestNum uint64, targetBlockNum uint64) {
			stable := true
			if targetBlockNum > latestNum-uint64(confirmedNum) {
				stable = false
			}

			block, transactions, err := bu.FetchBlock(ctx, big.NewInt(int64(targetBlockNum)), stable)
			if err != nil {
				log.Err(err).Msg("Fetch block fail when sync to latest block")
			}
			//no more block
			if block == nil {
				return
			}

			for _, transaction := range transactions {
				go bu.transactionUcase.SaveTransaction(ctx, block.BlockHash, transaction)
			}

			//it may cause some duplicate key err, but i don't think it would be a issue
			err = bu.newBlock(ctx, block)
			if err != nil {
				log.Err(err).Msg("New block fail when sync to latest block")
			}

			//job done , and release worker
			<-c
		}(latestNum, targetBlockNum.Uint64())
		targetBlockNum = targetBlockNum.Add(targetBlockNum, big.NewInt(1))
	}

}

func (bu *blockUseCase) setNewBlock(ctx context.Context, blockHash common.Hash) {
	block, err := bu.wsClient.BlockByHash(context.Background(), blockHash)
	if err != nil {
		log.Error().Err(err)
	}

	//store new block info
	if err := bu.newBlock(ctx, wrapBlockDb(block, false)); err != nil {
		log.Error().Err(err)
	}
}

//setOldBlock set old block to stable
func (bu *blockUseCase) setOldBlock(ctx context.Context, oldBlockNum uint64) {
	err := bu.setBlockStable(ctx, oldBlockNum, true)
	if err != nil && err != domain.ErrBlockNotExist {
		log.Error().Err(err)
	} else if err == domain.ErrBlockNotExist {
		b, transactions, err := bu.FetchBlock(ctx, big.NewInt(int64(oldBlockNum)), true)
		if err != nil {
			log.Error().Err(err)
		}

		for _, transaction := range transactions {
			go bu.transactionUcase.SaveTransaction(ctx, b.BlockHash, transaction)
		}

		err = bu.newBlock(ctx, b)
		if err != nil {
			log.Error().Err(err)
		}

	}
	return
}

func (bu *blockUseCase) subscribeNewBlock(ctx context.Context) {
	headers := make(chan *types.Header)

	sub, err := bu.wsClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Error().Err(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Error().Err(err)
		case header := <-headers: //get new block event
			fmt.Println("get new block:", header.Number)
			go bu.setNewBlock(ctx, header.Hash())
			go bu.setOldBlock(ctx, header.Number.Uint64()-uint64(confirmedNum))
		case <-ctx.Done():
			log.Print("break subscribe loop")
			return
		}
	}
}

func wrapBlockDb(block *types.Block, stable bool) *domain.BlockDb {
	return &domain.BlockDb{
		BlockNum:   block.NumberU64(),
		BlockHash:  block.Hash().String(),
		BlockTime:  block.Time(),
		ParentHash: block.ParentHash().String(),
		Stable:     stable,
	}
}

func (bu *blockUseCase) FetchBlock(ctx context.Context, blockNum *big.Int, stable bool) (*domain.BlockDb, types.Transactions, error) {
	block, err := bu.rpcClient.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		log.Err(err).Msg("fetch block fail in block by number")
		return nil, nil, err
	}

	if block == nil {
		return nil, nil, nil
	}
	log.Info().Msg("fetch block =" + block.Number().String())

	return wrapBlockDb(block, stable), block.Transactions(), nil
}
