package router

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/types"
)

// router bridges
var (
	RouterBridges map[string]tokens.CrossChainBridge // key is chainID
)

// Bridge eth bridge
type Bridge struct {
	*tokens.CrossChainBridgeBase
	*NonceSetterBase
	Signer types.Signer
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	return &Bridge{
		CrossChainBridgeBase: tokens.NewCrossChainBridgeBase(isSrc),
		NonceSetterBase:      NewNonceSetterBase(),
	}
}

// GetBridgeByChainID get bridge by chain id
func GetBridgeByChainID(chainID string) tokens.CrossChainBridge {
	return RouterBridges[chainID]
}

// InitRouterBridges init router bridges
func InitRouterBridges(isServer bool) {
}
