package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/domain"
	"github.com/ryanCool/ethService/helper"
	"net/http"
)

// TransactionHandler  represent the httphandler for article
type TransactionHandler struct {
	TUseCase domain.TransactionUseCase
}

func NewTransactionHandler(e *gin.Engine, dt domain.TransactionUseCase) {
	handler := &TransactionHandler{
		TUseCase: dt,
	}

	dg := e.Group("transaction")

	dg.GET("/:txHash", handler.GetTransaction)
}

func (a *TransactionHandler) GetTransaction(ctx *gin.Context) {
	txHash := ctx.Param("txHash")

	transaction, err := a.TUseCase.GetByTxHash(ctx, txHash)
	if err == domain.ErrTransactionNotExist {
		helper.RespondWithError(ctx, http.StatusNotFound, err)
	}

	if err != nil {
		fmt.Println(err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, transaction)
}
