package usecase

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/domain"
	"gorm.io/gorm"
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
var confirmedNum, scanWorkerNum, writeTransactionWorkerNum int

//Initialize init cron job to subscribe new block event through websocket endpoint
func (bu *blockUseCase) Initialize(ctx context.Context) {
	confirmedNum = config.GetInt("CONFIRMED_BLOCK_NUM")
	scanWorkerNum = config.GetInt("SCAN_WORK_NUM")
	syncFromNBlock = config.GetBigInt("SYNC_BLOCK_FROM_N")
	writeTransactionWorkerNum = config.GetInt("WRITE_TRANSACTION_WORK_NUM")

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
	log.Info().Uint64("latestNum", latestNum).Msg("latest block num=")
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

			err = bu.saveBlock(ctx, targetBlockNum, stable)
			if err != nil {
				log.Err(err).Msg("save block fail")
			}

			//job done , and release worker
			<-c
		}(latestNum, targetBlockNum.Uint64())
		targetBlockNum = targetBlockNum.Add(targetBlockNum, big.NewInt(1))
	}

}

func (bu *blockUseCase) saveBlock(ctx context.Context, blockNum uint64, stable bool) error {
	//check exist block in db
	b, err := bu.repo.GetByNumber(ctx, blockNum)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Err(err).Msg("get block from repo fail")
		return err
	}

	//block exist , and is not stable block  . Don't need to replace
	if b != nil && !stable {
		return nil
	} else if b != nil && stable { //exist old , we should replace by new fetch one . Delete first
		err = bu.repo.DeleteByNum(ctx, blockNum)
		if err != nil {
			log.Err(err).Msg("delete block by num fail")
			return err
		}
	}

	//block not exist in db
	block, transactions, err := bu.FetchBlock(ctx, big.NewInt(int64(blockNum)), stable)
	if err != nil {
		log.Err(err).Msg("Fetch block fail when sync to latest block")
		return err
	}

	err = bu.repo.Create(ctx, block)
	if err != nil {
		log.Err(err).Msg("New block fail when sync to latest block")
		return err
	}

	c := make(chan bool, writeTransactionWorkerNum)
	for _, transaction := range transactions {
		c <- true
		go func() {
			err = bu.transactionUcase.Save(ctx, block.BlockHash, transaction)
			if err != nil {
				log.Err(err).Msg("save transaction fail when sync to latest block")
			}
		}()
		<-c
	}
	return nil
}

func (bu *blockUseCase) setNewBlock(ctx context.Context, blockNum uint64) {
	err := bu.saveBlock(ctx, blockNum, false)
	if err != nil {
		log.Err(err).Msg("set new block fail - saveBlock")
	}
}

//setOldBlock set old block to stable
func (bu *blockUseCase) setOldBlock(ctx context.Context, oldBlockNum uint64) {
	b, err := bu.rpcClient.BlockByNumber(ctx, big.NewInt(int64(oldBlockNum)))
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail - Get by number")
	}

	oldBlock, err := bu.GetByNumber(ctx, oldBlockNum)
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail - Get by number")
	}

	//replace if old one is unstable by checking block hash
	//if block not exist or hash not equal , new one
	if oldBlock == nil || oldBlock.BlockHash != b.Hash().String() {
		err = bu.saveBlock(ctx, oldBlockNum, true)
		if err != nil {
			log.Error().Err(err).Msg("save block fail in set old block")
		}
		return
	}

	err = bu.setBlockStable(ctx, oldBlockNum, true)
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail")
	}

	return
}

func (bu *blockUseCase) subscribeNewBlock(ctx context.Context) {
	headers := make(chan *types.Header)
	sub, err := bu.wsClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Error().Err(err).Msg("subscribe new block fail")
	}

	for {
		select {
		case err := <-sub.Err():
			log.Error().Err(err).Msg("receive subsc")
		case header := <-headers: //get new block event
			log.Info().Uint64("block_num", header.Number.Uint64()).Msg("get new block")
			go bu.setNewBlock(ctx, header.Number.Uint64())
			go bu.setOldBlock(ctx, header.Number.Uint64()-uint64(confirmedNum))
		case <-ctx.Done():
			log.Print("break subscribe loop")
			return
		}
	}
}

func (bu *blockUseCase) FetchBlock(ctx context.Context, blockNum *big.Int, stable bool) (*domain.BlockDb, types.Transactions, error) {
	block, err := bu.rpcClient.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		log.Err(err).Msg("fetch block fail in block by number")
		return nil, nil, err
	}

	log.Info().Msg("fetch block =" + block.Number().String())

	return wrapBlockDb(block, stable), block.Transactions(), nil
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

func wrapBlockDb(block *types.Block, stable bool) *domain.BlockDb {
	return &domain.BlockDb{
		BlockNum:   block.NumberU64(),
		BlockHash:  block.Hash().String(),
		BlockTime:  block.Time(),
		ParentHash: block.ParentHash().String(),
		Stable:     stable,
	}
}
