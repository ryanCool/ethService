package domain

import (
	"fmt"
)

var (
	ErrBlockNotExist       = fmt.Errorf("block not exist")
	ErrTransactionNotExist = fmt.Errorf("transaction not exist")
)

var ErrMap = map[error]ErrCode{
	ErrBlockNotExist:       1001,
	ErrTransactionNotExist: 2001,
}

type ErrorResponse struct {
	ErrCode ErrCode `json:"err_code"`
	ErrMsg  ErrMsg  `json:"err_msg"`
}

type ErrCode int32

type ErrMsg string

var ErrMsgMap = map[ErrCode]ErrMsg{
	1001: "block not exist",
	2001: "transaction not exist",
}
