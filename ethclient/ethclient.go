package ethclient

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ryanCool/ethService/config"
)

var endpointURL, wsEndpointURL string
var (
	WsClient  *ethclient.Client
	RpcClient *ethclient.Client
)

func init() {
	// todo use multiple endpoint to avoid 429 too many request
	endpointURL = config.GetString("JSON_RPC_ENDPOINT")
	wsEndpointURL = config.GetString("WS_ENDPOINT")
}

func Initialize() {
	var err error
	RpcClient, err = ethclient.Dial(endpointURL)
	if err != nil {
		panic(err)
	}

	WsClient, err = ethclient.Dial(wsEndpointURL)
	if err != nil {
		panic(err)
	}
}

func Finalize() {
	RpcClient.Close()
	WsClient.Close()
}
