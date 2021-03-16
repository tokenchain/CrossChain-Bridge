package router

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/dcrm"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/types"
)

// router bridges
var (
	RouterBridges = make(map[string]*Bridge)           // key is chainID
	PeerTokens    = make(map[string]map[string]string) // key is tokenID,chainID
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
		bridge := NewCrossChainBridge()
		bridge.initGatewayConfig(chainID)
		bridge.initChainConfig(chainID)

		for _, tokenID := range tokenIDs {
			bridge.initTokenConfig(tokenID, chainID)
		}

		RouterBridges[chainID.String()] = bridge
	}
	printPeerTokens()

	cfg := params.GetRouterConfig()
	dcrm.Init(cfg.Dcrm, isServer)

	log.Info(">>> init router bridges success", "isServer", isServer)
}

func (b *Bridge) initGatewayConfig(chainID *big.Int) {
	if chainID.Sign() == 0 {
		log.Fatal("zero chain ID")
	}
	cfg := params.GetRouterConfig()
	apiAddrs := cfg.Gateways[chainID.String()]
	if len(apiAddrs) == 0 {
		log.Fatal("gateway not found for chain ID", "chainID", chainID)
	}
	b.SetGatewayConfig(&GatewayConfig{
		APIAddress: apiAddrs,
	})
	log.Infof(">>> [%5v] init gateway config success", chainID)
}

func (b *Bridge) initChainConfig(chainID *big.Int) bool {
	chainCfg, err := GetChainConfig(chainID)
	if err != nil {
		log.Fatal("get chain config failed", "chainID", chainID, "err", err)
	}
	if chainCfg == nil {
		log.Fatal("chain config not found", "chainID", chainID)
	}
	if chainID.String() != chainCfg.ChainID {
		log.Fatal("verify chain ID mismatch", "inconfig", chainCfg.ChainID, "inchainids", chainID)
	}
	b.SetChainConfig(chainCfg)
	log.Infof(">>> [%5v] init chain config success", chainID)
	return true
}

func (b *Bridge) initTokenConfig(tokenID string, chainID *big.Int) {
	if tokenID == "" {
		log.Fatal("empty token ID")
	}
	tokenAddr, err := GetPeerTokenOnChain(tokenID, chainID)
	if err != nil {
		log.Fatal("get token address failed", "tokenID", tokenID, "chainID", chainID, "err", err)
	}
	if common.HexToAddress(tokenAddr) == (common.Address{}) {
		log.Warnf(">>> [%5v] '%v' token address is empty", chainID, tokenID)
		return
	}
	tokenCfg, err := GetTokenConfig(chainID, tokenAddr)
	if err != nil {
		log.Fatal("get token config failed", "chainID", chainID, "tokenAddr", tokenAddr, "err", err)
	}
	if tokenCfg == nil {
		log.Warn("token config not found", "tokenID", tokenID, "chainID", chainID, "tokenAddr", tokenAddr)
		return
	}
	if common.HexToAddress(tokenAddr) != common.HexToAddress(tokenCfg.ContractAddress) {
		log.Fatal("verify token address mismach", "tokenID", tokenID, "chainID", chainID, "inconfig", tokenCfg.ContractAddress, "inpeer", tokenAddr)
	}
	if tokenID != tokenCfg.ID {
		log.Fatal("verify token ID mismatch", "chainID", chainID, "inconfig", tokenCfg.ID, "intokenids", tokenID)
	}
	b.SetTokenConfig(tokenAddr, tokenCfg)
	log.Info(fmt.Sprintf(">>> [%5v] init '%v' token config success", chainID, tokenID), "tokenAddr", tokenAddr, "decimals", tokenCfg.Decimals)

	tokenIDKey := strings.ToLower(tokenID)
	tokensMap := PeerTokens[tokenIDKey]
	if tokensMap == nil {
		tokensMap = make(map[string]string)
		PeerTokens[tokenIDKey] = tokensMap
	}
	tokensMap[chainID.String()] = tokenAddr
}

func printPeerTokens() {
	log.Info(">>> begin print all peer tokens")
	for tokenID, tokensMap := range PeerTokens {
		log.Infof(">>> peer tokens of tokenID '%v' count is %v", tokenID, len(tokensMap))
		for chainID, tokenAddr := range tokensMap {
			log.Infof(">>> peer token: chainID %v tokenAddr %v", chainID, tokenAddr)
		}
	}
	log.Info(">>> end print all peer tokens")
}
