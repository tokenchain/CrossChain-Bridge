package mongodb

import (
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/log"
)

func getVaultSwapKey(fromChainID, txid string, logindex int) string {
	return fmt.Sprintf("%v:%v:%v", fromChainID, txid, logindex)
}

// AddVaultSwap add swapout
func AddVaultSwap(ms *MgoSwap) error {
	ms.Key = getVaultSwapKey(ms.FromChainID, ms.TxID, ms.LogIndex)
	err := collVaultSwap.Insert(ms)
	if err == nil {
		log.Info("mongodb add vault swap success", "chainid", ms.FromChainID, "txid", ms.TxID, "logindex", ms.LogIndex)
	} else {
		log.Debug("mongodb add vault swap failed", "chainid", ms.FromChainID, "txid", ms.TxID, "logindex", ms.LogIndex, "err", err)
	}
	return mgoError(err)
}

// UpdateVaultSwapStatus update swapout status
func UpdateVaultSwapStatus(fromChainID, txid string, logindex int, status SwapStatus, timestamp int64, memo string) error {
	return nil
}

// FindVaultSwap api
func FindVaultSwap(fromChainID, txid string, logindex int) (*MgoSwap, error) {
	return nil, nil
}

// FindVaultSwapsWithStatus find swapout with status
func FindVaultSwapsWithStatus(status SwapStatus, septime int64) ([]*MgoSwap, error) {
	return nil, nil
}

// FindVaultSwapsWithChainIDAndStatus find swapout with fromChainID and status in the past septime
func FindVaultSwapsWithChainIDAndStatus(fromChainID string, status SwapStatus, septime int64) ([]*MgoSwap, error) {
	return nil, nil
}

// FindVaultSwapResult api
func FindVaultSwapResult(fromChainID, txid string, logindex int) (*MgoSwapResult, error) {
	return nil, nil
}

// FindVaultSwapResults api
func FindVaultSwapResults(fromChainID, address string, offset, limit int) ([]*MgoSwapResult, error) {
	return nil, nil
}
