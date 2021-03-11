package router

import (
	"math/big"
	"strings"
)

// CrossChainBridgeBase base bridge
type CrossChainBridgeBase struct {
	ChainConfig    *ChainConfig
	GatewayConfig  *GatewayConfig
	TokenConfigMap map[string]*TokenConfig
}

// NewCrossChainBridgeBase new base bridge
func NewCrossChainBridgeBase() *CrossChainBridgeBase {
	return &CrossChainBridgeBase{
		TokenConfigMap: make(map[string]*TokenConfig),
	}
}

// SetChainConfig set chain config
func (b *CrossChainBridgeBase) SetChainConfig(chainCfg *ChainConfig) {
	b.ChainConfig = chainCfg
}

// SetGatewayConfig set gateway config
func (b *CrossChainBridgeBase) SetGatewayConfig(gatewayCfg *GatewayConfig) {
	b.GatewayConfig = gatewayCfg
}

// SetTokenConfig set token config
func (b *CrossChainBridgeBase) SetTokenConfig(token string, tokenCfg *TokenConfig) {
	b.TokenConfigMap[strings.ToLower(token)] = tokenCfg
}

// GetChainConfig get chain config
func (b *CrossChainBridgeBase) GetChainConfig() *ChainConfig {
	return b.ChainConfig
}

// GetGatewayConfig get gateway config
func (b *CrossChainBridgeBase) GetGatewayConfig() *GatewayConfig {
	return b.GatewayConfig
}

// GetTokenConfig get token config
func (b *CrossChainBridgeBase) GetTokenConfig(token string) *TokenConfig {
	return b.TokenConfigMap[strings.ToLower(token)]
}

// GetBigValueThreshold get big value threshold
func (b *CrossChainBridgeBase) GetBigValueThreshold(token string) *big.Int {
	return b.GetTokenConfig(token).GetBigValueThreshold()
}

// CheckTokenSwapValue check swap value is in right range
func CheckTokenSwapValue(token *TokenConfig, value *big.Int) bool {
	if value == nil {
		return false
	}
	if value.Cmp(token.minSwap) < 0 {
		return false
	}
	if value.Cmp(token.maxSwap) > 0 {
		return false
	}
	swappedValue := CalcSwapValue(token, value)
	return swappedValue.Sign() > 0
}

// CalcSwapValue calc swap value (get rid of fee)
func CalcSwapValue(token *TokenConfig, value *big.Int) *big.Int {
	if token.SwapFeeRate == 0.0 {
		return value
	}

	feeRateMul1e18 := new(big.Int).SetUint64(uint64(token.SwapFeeRate * 1e18))
	swapFee := new(big.Int).Mul(value, feeRateMul1e18)
	swapFee.Div(swapFee, big.NewInt(1e18))

	if swapFee.Cmp(token.minSwapFee) < 0 {
		swapFee = token.minSwapFee
	} else if swapFee.Cmp(token.maxSwapFee) > 0 {
		swapFee = token.maxSwapFee
	}

	if value.Cmp(swapFee) > 0 {
		return new(big.Int).Sub(value, swapFee)
	}
	return big.NewInt(0)
}
