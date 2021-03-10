package router

import (
	"github.com/anyswap/CrossChain-Bridge/types"
)

// router bridges
var (
	RouterBridges map[string]*Bridge // key is chainID
)

// Bridge eth bridge
type Bridge struct {
	*CrossChainBridgeBase
	*NonceSetterBase
	Signer types.Signer
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge() *Bridge {
	return &Bridge{
		CrossChainBridgeBase: NewCrossChainBridgeBase(),
		NonceSetterBase:      NewNonceSetterBase(),
	}
}

// GetBridgeByChainID get bridge by chain id
func GetBridgeByChainID(chainID string) *Bridge {
	return RouterBridges[chainID]
}

// InitRouterBridges init router bridges
func InitRouterBridges(isServer bool) {
}
