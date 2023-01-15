package usecase

import (
	"context"
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
var latestBlockNum uint64
var confirmedNum int

func init() {
	confirmedNum = config.GetInt("CONFIRMED_BLOCK_NUM")

	syncFromNBlock = config.GetBigInt("SYNC_BLOCK_FROM_N")
}

//Initialize init cron job to subscribe new block event through websocket endpoint
func (bu *blockUseCase) Initialize(ctx context.Context) {
	header, err := bu.rpcClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	latestBlockNum = header.Number.Uint64()
	log.Println("latestBlockNum=", latestBlockNum)

	go bu.scanToLatest(ctx)
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

//List list latest limit block
func (bu *blockUseCase) List(ctx context.Context, limit int) ([]domain.BlockDb, error) {
	return bu.repo.List(ctx, limit)
}

//ScanToLatest scan blocks from n to latest , and store to db
func (bu *blockUseCase) scanToLatest(ctx context.Context) {
	currentBlockNum := syncFromNBlock

	for {
		block, err := bu.FetchBlock(ctx, currentBlockNum)
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

func (bu *blockUseCase) FetchBlock(ctx context.Context, blockNum *big.Int) (*domain.BlockDb, error) {
	block, err := bu.rpcClient.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if block == nil {
		return nil, nil
	}
	log.Println("fetch block =", block.NumberU64())

	stable := true
	if block.NumberU64() > latestBlockNum-uint64(confirmedNum) {
		stable = false
	}

	return &domain.BlockDb{
		BlockNum:   block.NumberU64(),
		BlockHash:  block.Hash().String(),
		BlockTime:  block.Time(),
		ParentHash: block.ParentHash().String(),
		Stable:     stable,
	}, nil

}
