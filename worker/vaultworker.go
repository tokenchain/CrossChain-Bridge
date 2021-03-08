package worker

import (
	"time"
)

// StartVaultSwapWork start vault swap job
func StartVaultSwapWork(isServer bool) {
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
