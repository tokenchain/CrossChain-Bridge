package worker

import (
	"time"
)

// StartRouterSwapWork start router swap job
func StartRouterSwapWork(isServer bool) {
	if !isServer {
		go StartAcceptSignJob()
		return
	}

	go StartVerifyJob()
	time.Sleep(interval)

	go StartSwapJob()
	time.Sleep(interval)

	go StartStableJob()
	time.Sleep(interval)

	go StartReplaceJob()
}
