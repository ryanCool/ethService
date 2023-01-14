package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/database"
	"github.com/ryanCool/ethService/ethclient"
	"net/http"
)

func main() {
	ctx := context.Background()

	ethclient.New()
	database.Initialize(ctx)
	defer database.Finalize(ctx)
	engine := gin.New()

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: engine,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
