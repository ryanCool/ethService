package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/domain"
	"github.com/ryanCool/ethService/helper"
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

	if iLimit > 100 {
		ctx.JSON(http.StatusBadRequest, "limit should be 0~100")
		return
	}

	results, err := a.BUseCase.List(ctx, iLimit)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{"blocks": results})
}

func (a *BlockHandler) GetBlock(ctx *gin.Context) {
	blockNum := ctx.Param("id")
	iBlockNum, err := strconv.Atoi(blockNum)
	if err != nil {
		fmt.Println(err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	block, err := a.BUseCase.GetByNumber(ctx, uint64(iBlockNum))
	if err == domain.ErrBlockNotExist {
		helper.RespondWithError(ctx, http.StatusNotFound, err)
	}

	if err != nil {
		fmt.Println(err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, block)
}
