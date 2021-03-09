package mongodb

import (
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/log"
)

func getRouterSwapKey(fromChainID, txid string, logindex int) string {
	return fmt.Sprintf("%v:%v:%v", fromChainID, txid, logindex)
}

// AddRouterSwap add swapout
func AddRouterSwap(ms *MgoSwap) error {
	ms.Key = getRouterSwapKey(ms.FromChainID, ms.TxID, ms.LogIndex)
	err := collRouterSwap.Insert(ms)
	if err == nil {
		log.Info("mongodb add router swap success", "chainid", ms.FromChainID, "txid", ms.TxID, "logindex", ms.LogIndex)
	} else {
		log.Debug("mongodb add router swap failed", "chainid", ms.FromChainID, "txid", ms.TxID, "logindex", ms.LogIndex, "err", err)
	}
	return mgoError(err)
}

// UpdateRouterSwapStatus update swapout status
func UpdateRouterSwapStatus(fromChainID, txid string, logindex int, status SwapStatus, timestamp int64, memo string) error {
	return nil
}

// FindRouterSwap api
func FindRouterSwap(fromChainID, txid string, logindex int) (*MgoSwap, error) {
	return nil, nil
}

// FindRouterSwapsWithStatus find swapout with status
func FindRouterSwapsWithStatus(status SwapStatus, septime int64) ([]*MgoSwap, error) {
	return nil, nil
}

// FindRouterSwapsWithChainIDAndStatus find swapout with fromChainID and status in the past septime
func FindRouterSwapsWithChainIDAndStatus(fromChainID string, status SwapStatus, septime int64) ([]*MgoSwap, error) {
	return nil, nil
}

// FindRouterSwapResult api
func FindRouterSwapResult(fromChainID, txid string, logindex int) (*MgoSwapResult, error) {
	return nil, nil
}

// FindRouterSwapResults api
func FindRouterSwapResults(fromChainID, address string, offset, limit int) ([]*MgoSwapResult, error) {
	return nil, nil
}
