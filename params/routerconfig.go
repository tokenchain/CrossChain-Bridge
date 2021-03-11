package params

import (
	"encoding/json"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
)

// router swap constants
const (
	RouterSwapIdentifier = "routerswap"
)

var (
	routerConfig *RouterConfig
)

// RouterServerConfig only for server
type RouterServerConfig struct {
	Admins    []string         `toml:",omitempty" json:",omitempty"`
	MongoDB   *MongoDBConfig   `toml:",omitempty" json:",omitempty"`
	APIServer *APIServerConfig `toml:",omitempty" json:",omitempty"`
}

// RouterConfig config
type RouterConfig struct {
	*RouterServerConfig `toml:",omitempty" json:",omitempty"`

	Identifier string
	Onchain    *OnchainConfig `toml:",omitempty" json:",omitempty"`
	Dcrm       *DcrmConfig    `toml:",omitempty" json:",omitempty"`
}

// OnchainConfig struct
type OnchainConfig struct {
	Contract   string
	APIAddress []string
}

// GetRouterConfig get router config
func GetRouterConfig() *RouterConfig {
	return routerConfig
}

// HasRouterAdmin has admin
func HasRouterAdmin() bool {
	return len(routerConfig.Admins) != 0
}

// IsRouterAdmin is admin
func IsRouterAdmin(account string) bool {
	for _, admin := range routerConfig.Admins {
		if strings.EqualFold(account, admin) {
			return true
		}
	}
	return false
}

// IsRouterSwap is router swap
func IsRouterSwap() bool {
	return strings.EqualFold(routerConfig.Identifier, RouterSwapIdentifier)
}

// LoadRouterConfig load router swap config
func LoadRouterConfig(configFile string, isServer bool) *RouterConfig {
	if configFile == "" {
		log.Fatal("must specify config file")
	}
	log.Info("load router config file", "configFile", configFile, "isServer", isServer)
	if !common.FileExist(configFile) {
		log.Fatalf("LoadRouterConfig error: config file '%v' not exist", configFile)
	}
	config := &RouterConfig{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("LoadRouterConfig error (toml DecodeFile): %v", err)
	}

	if !isServer {
		config.RouterServerConfig = nil
	}

	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Println("LoadRouterConfig finished.", string(bs))
	if err := config.CheckConfig(isServer); err != nil {
		log.Fatalf("Check config failed. %v", err)
	}

	routerConfig = config
	return routerConfig
}
