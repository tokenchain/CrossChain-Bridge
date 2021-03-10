package params

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// router swap constants
const (
	RouterSwapIdentifier = "routerswap"
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

// IsRouterSwap is router swap
func IsRouterSwap() bool {
	return strings.EqualFold(GetIdentifier(), RouterSwapIdentifier)
}

// LoadRouterConfig load router swap config
func LoadRouterConfig(configFile string, isServer bool) *ServerConfig {
	log.Info("load router config file", "configFile", configFile, "isServer", isServer)
	if !common.FileExist(configFile) {
		log.Fatalf("LoadRouterConfig error: config file '%v' not exist", configFile)
	}
	config := &ServerConfig{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("LoadRouterConfig error (toml DecodeFile): %v", err)
	}

	SetConfig(config)
	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Println("LoadRouterConfig finished.", string(bs))
	if err := CheckConfig(isServer); err != nil {
		log.Fatalf("Check config failed. %v", err)
	}
	return serverConfig
}
