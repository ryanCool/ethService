package eth

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/domain"
	"math/big"
)

var syncFromNBlock *big.Int
var confirmedNum, scanWorkerNum, writeTransactionWorkerNum int

//Initialize init cron job to subscribe new block event through websocket endpoint
func (es *ethScan) Initialize(ctx context.Context) {
	confirmedNum = config.GetInt("CONFIRMED_BLOCK_NUM")
	scanWorkerNum = config.GetInt("SCAN_WORK_NUM")
	syncFromNBlock = config.GetBigInt("SYNC_BLOCK_FROM_N")
	writeTransactionWorkerNum = config.GetInt("WRITE_TRANSACTION_WORK_NUM")

	go es.subscribeNewBlock(ctx)
	//go es.scanToLatest(ctx)
}

type ethScan struct {
	rpcClient        *ethclient.Client
	wsClient         *ethclient.Client
	transactionUcase domain.TransactionUseCase
	blockUCase       domain.BlockUseCase
}

func NewEthScan(rpcClient *ethclient.Client, wsClient *ethclient.Client, transactionUcase domain.TransactionUseCase, blockUcase domain.BlockUseCase) ethScan {
	return ethScan{
		rpcClient:        rpcClient,
		wsClient:         wsClient,
		transactionUcase: transactionUcase,
		blockUCase:       blockUcase,
	}
}

func (es *ethScan) subscribeNewBlock(ctx context.Context) {
	headers := make(chan *types.Header)
	sub, err := es.wsClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Error().Err(err).Msg("subscribe new block fail")
	}

	for {
		select {
		case err := <-sub.Err():
			log.Error().Err(err).Msg("receive subsc")
		case header := <-headers: //get new block event
			log.Info().Uint64("block_num", header.Number.Uint64()).Msg("get new block")
			go es.setNewBlock(ctx, header.Number.Uint64())
			go es.setOldBlock(ctx, header.Number.Uint64()-uint64(confirmedNum))
		case <-ctx.Done():
			log.Print("break subscribe loop")
			return
		}
	}
}

func (es *ethScan) FetchBlock(ctx context.Context, blockNum *big.Int, stable bool) (*domain.BlockDb, types.Transactions, error) {
	block, err := es.rpcClient.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		log.Err(err).Msg("fetch block fail in block by number")
		return nil, nil, err
	}

	log.Info().Msg("fetch block =" + block.Number().String())

	return wrapBlockDb(block, stable), block.Transactions(), nil
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

func (es *ethScan) setNewBlock(ctx context.Context, blockNum uint64) {
	err := es.saveBlock(ctx, blockNum, false)
	if err != nil {
		log.Err(err).Msg("set new block fail - saveBlock")
	}
}

//setOldBlock set old block to stable
func (es *ethScan) setOldBlock(ctx context.Context, oldBlockNum uint64) {
	b, err := es.rpcClient.BlockByNumber(ctx, big.NewInt(int64(oldBlockNum)))
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail - Get by number")
	}

	oldBlock, err := es.blockUCase.GetByNumber(ctx, oldBlockNum)
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail - Get by number")
	}

	//replace if old one is unstable by checking block hash
	//if block not exist or hash not equal , new one
	if oldBlock == nil || oldBlock.BlockHash != b.Hash().String() {
		err = es.saveBlock(ctx, oldBlockNum, true)
		if err != nil {
			log.Error().Err(err).Msg("save block fail in set old block")
		}
		return
	}

	err = es.setBlockStable(ctx, oldBlockNum, true)
	if err != nil {
		log.Error().Err(err).Msg("set block stable fail")
	}

	return
}

func (es *ethScan) saveBlock(ctx context.Context, blockNum uint64, stable bool) error {
	//check exist block in db
	b, err := es.blockUCase.GetByNumber(ctx, blockNum)
	if err != nil && err != domain.ErrBlockNotExist {
		log.Err(err).Msg("get block from blockRepo fail")
		return err
	}

	//block exist , and is not stable block  . Don't need to replace
	if b != nil && !stable {
		return nil
	} else if b != nil && stable { //exist old , we should replace by new fetch one . Delete first
		err = es.blockUCase.DeleteByNum(ctx, blockNum)
		if err != nil {
			log.Err(err).Msg("delete block by num fail")
			return err
		}
	}

	//block not exist in db
	block, transactions, err := es.FetchBlock(ctx, big.NewInt(int64(blockNum)), stable)
	if err != nil {
		log.Err(err).Msg("Fetch block fail when sync to latest block")
		return err
	}

	err = es.blockUCase.Create(ctx, block)
	if err != nil {
		log.Err(err).Msg("New block fail when sync to latest block")
		return err
	}

	c := make(chan bool, writeTransactionWorkerNum)
	for _, transaction := range transactions {
		c <- true
		go func(transaction types.Transaction) {
			err = es.Save(ctx, block.BlockHash, &transaction)
			if err != nil {
				log.Err(err).Msg("save transaction fail when sync to latest block")
			}
		}(*transaction)
		<-c
	}
	return nil
}

//ScanToLatest scan blocks from n to latest , and store to db
func (es *ethScan) scanToLatest(ctx context.Context) {
	header, err := es.rpcClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error().Err(err)
	}

	latestNum := header.Number.Uint64()
	log.Info().Uint64("latestNum", latestNum).Msg("latest block num=")
	targetBlockNum := syncFromNBlock

	//use esffer channel to implement a worker pool with config number
	c := make(chan bool, scanWorkerNum)
	for targetBlockNum.Uint64() <= latestNum {
		c <- true
		go func(latestNum uint64, targetBlockNum uint64) {
			stable := true
			if targetBlockNum > latestNum-uint64(confirmedNum) {
				stable = false
			}

			err = es.saveBlock(ctx, targetBlockNum, stable)
			if err != nil {
				log.Err(err).Msg("save block fail")
			}

			//job done , and release worker
			<-c
		}(latestNum, targetBlockNum.Uint64())
		targetBlockNum = targetBlockNum.Add(targetBlockNum, big.NewInt(1))
	}

}

//setBlockStable set old block to stable status
func (es *ethScan) setBlockStable(ctx context.Context, blockNum uint64, stable bool) error {
	return es.blockUCase.SetStable(ctx, blockNum, stable)
}

func (es *ethScan) saveReceipt(ctx context.Context, txHash common.Hash) error {
	receipt, err := es.rpcClient.TransactionReceipt(context.Background(), txHash)
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
		err = es.transactionUcase.SaveReceiptAndLogs(ctx, txHash.String(), logs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (es *ethScan) Save(ctx context.Context, blockHash string, transaction *types.Transaction) error {
	from, err := types.Sender(types.LatestSignerForChainID(transaction.ChainId()), transaction)
	if err != nil {
		return err
	}

	to := transaction.To()
	if to == nil {
		to = &common.Address{}
	}
	err = es.transactionUcase.Create(ctx, &domain.Transaction{
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
		err = es.saveReceipt(ctx, transaction.Hash())
		if err != nil {
			log.Err(err).Msg("save receipt fail")
		}
	}()

	return nil
}
