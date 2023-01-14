package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/domain"
	"net/http"
)

// BlockHandler  represent the httphandler for article
type BlockHandler struct {
	BUseCase domain.BlockUseCase
}

func NewBlockHandler(e *gin.Engine, ds domain.BlockUseCase) {
	handler := &BlockHandler{
		BUseCase: ds,
	}

	dg := e.Group("blocks")

	dg.GET("/:id", handler.GetBlock)
	dg.GET("/", handler.ListBlock)
}

func (a *BlockHandler) ListBlock(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNoContent)
}

func (a *BlockHandler) GetBlock(ctx *gin.Context) {

	ctx.AbortWithStatus(http.StatusNoContent)
}
