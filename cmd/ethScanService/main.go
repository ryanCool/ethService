package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	blockRepo "github.com/ryanCool/ethService/block/repository/postgres"
	blockUcase "github.com/ryanCool/ethService/block/usecase"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/database"
	"github.com/ryanCool/ethService/eth"
	"github.com/ryanCool/ethService/ethclient"
	transactionRepo "github.com/ryanCool/ethService/transaction/repository/postgres"
	transactionUcase "github.com/ryanCool/ethService/transaction/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create root context.
	ctx, cancel := context.WithCancel(context.Background())
	timeoutContext := time.Duration(config.GetInt("CONTEXT_TIMEOUT_SECS")) * time.Second

	ethclient.Initialize()
	defer ethclient.Finalize()

	database.Initialize(ctx)
	defer database.Finalize(ctx)

	db := database.GetDB()

	//init transaction service
	tp := transactionRepo.NewPostgresTransactionRepository(db)
	tu := transactionUcase.NewTransactionUseCase(tp, timeoutContext)

	//init block service
	bp := blockRepo.NewPostgresBlockRepository(db)
	bu := blockUcase.NewBlockUseCase(bp, tu, timeoutContext)

	ethScan := eth.NewEthScan(ethclient.RpcClient, ethclient.WsClient, tu, bu)
	ethScan.Initialize(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()

	log.Print("exit...")
}
