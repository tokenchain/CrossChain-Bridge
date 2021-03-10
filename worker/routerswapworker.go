package worker

import (
	"time"

	"github.com/anyswap/CrossChain-Bridge/rpc/client"
	"github.com/anyswap/CrossChain-Bridge/tokens/router"
	"github.com/anyswap/CrossChain-Bridge/worker/routerswap"
)

// StartRouterSwapWork start router swap job
func StartRouterSwapWork(isServer bool) {
	logWorker("worker", "start router swap worker")

	client.InitHTTPClient()
	router.InitRouterBridges(isServer)

	if !isServer {
		go routerswap.StartAcceptSignJob()
		return
	}

	go routerswap.StartVerifyJob()
	time.Sleep(interval)

	go routerswap.StartSwapJob()
	time.Sleep(interval)

	go routerswap.StartStableJob()
	time.Sleep(interval)

	go routerswap.StartReplaceJob()
}
