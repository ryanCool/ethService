package helper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ryanCool/ethService/domain"
)

// RespondWithError responds to the request with the provided error .
func RespondWithError(ctx *gin.Context, status int, err error) {

	// Compose error message for logging and response.
	message := err.Error()

	// Log error message and respond to request.
	fmt.Println(message)

	if v, exist := domain.ErrMap[err]; exist {
		ctx.AbortWithStatusJSON(status, domain.ErrorResponse{
			ErrCode: v,
			ErrMsg:  domain.ErrMsgMap[v],
		})
		return
	}

	ctx.AbortWithStatusJSON(status, gin.H{
		"error": message,
	})
}
