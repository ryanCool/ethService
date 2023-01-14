package ethclient

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

func New() *ethclient.Client {
	client, err := ethclient.Dial("https://data-seed-prebsc-2-s3.binance.org:8545/")
	if err != nil {
		panic(err)
	}

	return client
}
