package router

import (
	"github.com/anyswap/CrossChain-Bridge/dcrm"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/types"
)

// router bridges
var (
	RouterBridges = make(map[string]*Bridge) // key is chainID
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
	log.Info("start init router bridges")

	cfg := params.GetRouterConfig()

	chainIDs, err := GetAllChainIDs()
	if err != nil {
		log.Fatal("call GetAllChainIDs failed", "err", err)
	}
	log.Info("get all chain ids success", "chainIDs", chainIDs)

	tokenIDs, err := GetAllTokenIDs()
	if err != nil {
		log.Fatal("call GetAllTokenIDs failed", "err", err)
	}
	log.Info("get all token ids success", "tokenIDs", tokenIDs)

	for _, chainID := range chainIDs {
		chainCfg, errf := GetChainConfig(chainID)
		if errf != nil {
			log.Fatal("get chain config failed", "chainID", chainID, "err", errf)
		}
		if chainCfg == nil {
			log.Fatal("chain config not found", "chainID", chainID)
		}
		if chainID.String() != chainCfg.ChainID {
			log.Fatal("chain ID mismatch", "inconfig", chainCfg.ChainID, "inkey", chainID)
		}
		log.Info("get chain config success", "chain", chainCfg.BlockChain, "chainID", chainID)
		bridge := NewCrossChainBridge()
		bridge.SetChainConfig(chainCfg)
		RouterBridges[chainID.String()] = bridge
	}

	dcrm.Init(cfg.Dcrm, isServer)

	log.Info("init router bridges success", "isServer", isServer)
}
