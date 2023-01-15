package usecase

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/domain"
	"log"
	"math/big"
	"time"
)

type blockUseCase struct {
	repo           domain.BlockRepository
	rpcClient      *ethclient.Client
	wsClient       *ethclient.Client
	contextTimeout time.Duration
}

var syncFromNBlock *big.Int
var confirmedNum int

func init() {
	confirmedNum = config.GetInt("CONFIRMED_BLOCK_NUM")

	syncFromNBlock = config.GetBigInt("SYNC_BLOCK_FROM_N")
}

//Initialize init cron job to subscribe new block event through websocket endpoint
func (bu *blockUseCase) Initialize(ctx context.Context) {

	go bu.subscribeNewBlock(ctx)
	//go bu.scanToLatest(ctx)
}

func NewBlockUseCase(a domain.BlockRepository, timeout time.Duration, rpcClient *ethclient.Client, wsClient *ethclient.Client) domain.BlockUseCase {
	return &blockUseCase{
		repo:           a,
		contextTimeout: timeout,
		wsClient:       wsClient,
		rpcClient:      rpcClient,
	}
}

//NewBlock store block to db
func (bu *blockUseCase) newBlock(ctx context.Context, block *domain.BlockDb) error {
	return bu.repo.Create(ctx, block)
}

//NewBlock store block to db
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
		log.Fatal(err)
	}
	latestNum := header.Number.Uint64()
	currentBlockNum := syncFromNBlock

	for {
		stable := true
		if currentBlockNum.Uint64() > latestNum-uint64(confirmedNum) {
			stable = false
		}

		block, err := bu.FetchBlock(ctx, currentBlockNum, stable)
		if err != nil {
			log.Println(err)
		}
		if block == nil { //no more block
			break
		}

		err = bu.newBlock(ctx, block)
		if err != nil {
			log.Println(err)
		}
		currentBlockNum = currentBlockNum.Add(currentBlockNum, big.NewInt(1))
	}

}

func (bu *blockUseCase) subscribeNewBlock(ctx context.Context) {
	headers := make(chan *types.Header)

	sub, err := bu.wsClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := bu.wsClient.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			if err := bu.newBlock(ctx, wrapBlockDb(block, false)); err != nil {
				log.Fatal(err)
			}

			oldBlockNum := block.Number().Uint64() - uint64(confirmedNum)
			//set old block to stable
			err = bu.setBlockStable(ctx, oldBlockNum, true)
			if err != nil && err != domain.ErrBlockNotExist {
				log.Fatal(err)
			} else if err == domain.ErrBlockNotExist {
				b, err := bu.FetchBlock(ctx, big.NewInt(int64(oldBlockNum)), true)
				if err != nil {
					log.Fatal(err)
				}

				err = bu.newBlock(ctx, b)
				if err != nil {
					log.Fatal(err)
				}

			}

		case <-ctx.Done():
			log.Println("break subscribe loop")
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

func (bu *blockUseCase) FetchBlock(ctx context.Context, blockNum *big.Int, stable bool) (*domain.BlockDb, error) {
	block, err := bu.rpcClient.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if block == nil {
		return nil, nil
	}
	log.Println("fetch block =", block.NumberU64())

	return wrapBlockDb(block, stable), nil
}
