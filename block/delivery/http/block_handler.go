package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/domain"
	"net/http"
	"strconv"
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
	limit := ctx.Query("limit")
	iLimit, _ := strconv.Atoi(limit)
	if iLimit == 0 {
		//set default to 20
		iLimit = 20
	}

	results, err := a.BUseCase.List(ctx, iLimit)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{"blocks": results})
}

func (a *BlockHandler) GetBlock(ctx *gin.Context) {

	ctx.AbortWithStatus(http.StatusNoContent)
}
