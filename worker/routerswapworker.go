package worker

import (
	"time"

	"github.com/anyswap/CrossChain-Bridge/worker/routerswap"
)

// StartRouterSwapWork start router swap job
func StartRouterSwapWork(isServer bool) {
	if !isServer {
		go StartAcceptSignJob()
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
