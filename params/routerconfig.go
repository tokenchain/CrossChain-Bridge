package params

import (
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// OnchainConfig struct
type OnchainConfig struct {
	Gateway  tokens.GatewayConfig
	Contract string
}

// GetChainConfig impl
func GetChainConfig(chainID *big.Int) *tokens.ChainConfig {
	return nil
}

// GetTokenConfig impl
func GetTokenConfig(chainID *big.Int) *tokens.TokenConfig {
	return nil
}

// GetDcrmConfig impl
func GetDcrmConfig() *DcrmConfig {
	return nil
}
