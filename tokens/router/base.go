package router

import (
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// CrossChainBridgeBase base bridge
type CrossChainBridgeBase struct {
	ChainConfig    *tokens.ChainConfig
	GatewayConfig  *tokens.GatewayConfig
	TokenConfigMap map[string]*tokens.TokenConfig
}

// NewCrossChainBridgeBase new base bridge
func NewCrossChainBridgeBase() *CrossChainBridgeBase {
	return &CrossChainBridgeBase{
		TokenConfigMap: make(map[string]*tokens.TokenConfig),
	}
}

// SetChainConfig set chain config
func (b *CrossChainBridgeBase) SetChainConfig(chainCfg *tokens.ChainConfig) {
	b.ChainConfig = chainCfg
}

// SetGatewayConfig set gateway config
func (b *CrossChainBridgeBase) SetGatewayConfig(gatewayCfg *tokens.GatewayConfig) {
	b.GatewayConfig = gatewayCfg
}

// SetTokenConfig set token config
func (b *CrossChainBridgeBase) SetTokenConfig(token string, tokenCfg *tokens.TokenConfig) {
	b.TokenConfigMap[strings.ToLower(token)] = tokenCfg
}

// GetChainConfig get chain config
func (b *CrossChainBridgeBase) GetChainConfig() *tokens.ChainConfig {
	return b.ChainConfig
}

// GetGatewayConfig get gateway config
func (b *CrossChainBridgeBase) GetGatewayConfig() *tokens.GatewayConfig {
	return b.GatewayConfig
}

// GetTokenConfig get token config
func (b *CrossChainBridgeBase) GetTokenConfig(token string) *tokens.TokenConfig {
	return b.TokenConfigMap[strings.ToLower(token)]
}

// GetDcrmPublicKey get dcrm address's public key
func (b *CrossChainBridgeBase) GetDcrmPublicKey(token string) string {
	tokenCfg := b.GetTokenConfig(token)
	if tokenCfg != nil {
		return tokenCfg.DcrmPubkey
	}
	return ""
}

// GetBigValueThreshold get big value threshold
func (b *CrossChainBridgeBase) GetBigValueThreshold(token string) *big.Int {
	return b.GetTokenConfig(token).GetBigValueThreshold()
}

// CheckSwapValue check swap value is in right range
func (b *CrossChainBridgeBase) CheckSwapValue(token string, value *big.Int) bool {
	return tokens.CheckTokenSwapValue(b.GetTokenConfig(token), value)
}

// CalcSwapValue calc swap value (get rid of fee)
func (b *CrossChainBridgeBase) CalcSwapValue(token string, value *big.Int) *big.Int {
	return tokens.CalcSwapValue(b.GetTokenConfig(token), value)
}
