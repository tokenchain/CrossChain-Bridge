package mongodb

// AddVaultSwap add swapout
func AddVaultSwap(ms *MgoSwap) error {
	return nil
}

// UpdateVaultSwapStatus update swapout status
func UpdateVaultSwapStatus(fromChainID, txid, logindex string, status SwapStatus, timestamp int64, memo string) error {
	return nil
}

// FindVaultSwap api
func FindVaultSwap(fromChainID, txid, logindex string) (*MgoSwap, error) {
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
func FindVaultSwapResult(fromChainID, txid, logindex string) (*MgoSwapResult, error) {
	return nil, nil
}

// FindVaultSwapResults api
func FindVaultSwapResults(fromChainID, address string, offset, limit int) ([]*MgoSwapResult, error) {
	return nil, nil
}
