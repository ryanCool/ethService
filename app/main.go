package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	blockHttp "github.com/ryanCool/ethService/block/delivery/http"
	blockRepo "github.com/ryanCool/ethService/block/repository/postgres"
	blockUcase "github.com/ryanCool/ethService/block/usecase"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/database"
	"github.com/ryanCool/ethService/ethclient"
	transactionRepo "github.com/ryanCool/ethService/transaction/repository/postgres"
	transactionUcase "github.com/ryanCool/ethService/transaction/usecase"
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
	engine := gin.New()

	db := database.GetDB()

	//init transaction service
	tp := transactionRepo.NewPostgresTransactionRepository(db)
	tu := transactionUcase.NewTransactionUseCase(tp, timeoutContext, ethclient.RpcClient)

	//init block service
	bp := blockRepo.NewPostgresBlockRepository(db)
	bu := blockUcase.NewBlockUseCase(bp, tu, timeoutContext, ethclient.RpcClient, ethclient.WsClient)
	blockHttp.NewBlockHandler(engine, bu)

	bu.Initialize(ctx)

	//create http server to serve rest api
	serverAddress := fmt.Sprintf("%s:%s", config.GetString("SERVER_HOST"), config.GetString("SERVER_PORT"))
	server := &http.Server{
		Addr:    serverAddress,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	server.Close()

	log.Print("Shutdown Server ...")

}
