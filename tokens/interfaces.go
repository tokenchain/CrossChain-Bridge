package tokens

import (
	"math/big"
)

// CrossChainBridge interface
type CrossChainBridge interface {
	IsSrcEndpoint() bool

	SetChainAndGateway(*ChainConfig, *GatewayConfig)

	GetChainConfig() *ChainConfig
	GetGatewayConfig() *GatewayConfig
	GetTokenConfig(pairID string) *TokenConfig

	VerifyTokenConfig(*TokenConfig) error
	IsValidAddress(address string) bool

	GetTransaction(txHash string) (interface{}, error)
	GetTransactionStatus(txHash string) *TxStatus
	VerifyTransaction(pairID, txHash string, allowUnstable bool) (*TxSwapInfo, error)
	VerifyMsgHash(rawTx interface{}, msgHash []string) error

	BuildRawTransaction(args *BuildTxArgs) (rawTx interface{}, err error)
	SignTransaction(rawTx interface{}, pairID string) (signedTx interface{}, txHash string, err error)
	DcrmSignTransaction(rawTx interface{}, args *BuildTxArgs) (signedTx interface{}, txHash string, err error)
	SendTransaction(signedTx interface{}) (txHash string, err error)

	GetLatestBlockNumber() (uint64, error)
	GetLatestBlockNumberOf(apiAddress string) (uint64, error)
}

// ScanChainSupport interface
type ScanChainSupport interface {
	StartChainTransactionScanJob()
	StartPoolTransactionScanJob()
}

// ScanHistorySupport interface
type ScanHistorySupport interface {
	StartSwapHistoryScanJob()
}

// BalanceGetter interface
type BalanceGetter interface {
	GetBalance(accountAddress string) (*big.Int, error)
	GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error)
	GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error)
}

// NonceSetter interface (for eth-like)
type NonceSetter interface {
	GetPoolNonce(address, height string) (uint64, error)
	SetNonce(pairID string, value uint64)
	AdjustNonce(pairID string, value uint64) (nonce uint64)
	IncreaseNonce(pairID string, value uint64)
}
