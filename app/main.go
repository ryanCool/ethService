package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	blockHttp "github.com/ryanCool/ethService/block/delivery/http"
	blockRepo "github.com/ryanCool/ethService/block/repository/postgres"
	blockUcase "github.com/ryanCool/ethService/block/usecase"
	"github.com/ryanCool/ethService/config"
	"github.com/ryanCool/ethService/database"
	"github.com/ryanCool/ethService/ethclient"
)

func main() {
	ctx := context.Background()
	timeoutContext := time.Duration(config.GetInt("CONTEXT_TIMEOUT_SECS")) * time.Second

	ethclient.New()
	database.Initialize(ctx)
	defer database.Finalize(ctx)
	engine := gin.New()

	db := database.GetDB()
	bp := blockRepo.NewPostgresBlockRepository(db)
	bu := blockUcase.NewBlockUseCase(bp, timeoutContext)
	blockHttp.NewBlockHandler(engine, bu)

	serverAddress := fmt.Sprintf("%s:%s", config.GetString("SERVER_HOST"), config.GetString("SERVER_PORT"))
	server := &http.Server{
		Addr:    serverAddress,
		Handler: engine,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
