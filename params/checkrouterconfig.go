package params

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/common/hexutil"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/rpc/client"
)

// CallContractWithGateway call eth_call
func CallContractWithGateway(gateway, contract string, data hexutil.Bytes, blockNumber string) (result string, err error) {
	reqArgs := map[string]interface{}{
		"to":   contract,
		"data": data,
	}
	err = client.RPCPost(&result, gateway, "eth_call", reqArgs, blockNumber)
	if err == nil {
		return result, nil
	}
	return "", err
}

// CheckConfig check router config
func (config *RouterConfig) CheckConfig(isServer bool) (err error) {
	if config.Identifier != RouterSwapIdentifier {
		return fmt.Errorf("wrong identifier, have '%v', want '%v'", config.Identifier, RouterSwapIdentifier)
	}
	log.Info("check identifier pass", "identifier", config.Identifier, "isServer", isServer)
	if isServer {
		err = config.Server.CheckConfig()
		if err != nil {
			return err
		}
	}

	if config.Dcrm == nil {
		return errors.New("server must config 'Dcrm'")
	}
	err = config.Dcrm.CheckConfig(isServer)
	if err != nil {
		return err
	}

	if config.Onchain == nil {
		return errors.New("server must config 'Onchain'")
	}
	err = config.Onchain.CheckConfig()
	if err != nil {
		return err
	}

	return nil
}

// CheckConfig of router server
func (s *RouterServerConfig) CheckConfig() error {
	if s.MongoDB == nil {
		return errors.New("server must config 'MongoDB'")
	}
	if s.APIServer == nil {
		return errors.New("server must config 'APIServer'")
	}
	for _, chainID := range s.ChainIDBlackList {
		biChainID, ok := new(big.Int).SetString(chainID, 0)
		if !ok {
			return fmt.Errorf("wrong chain id '%v' in black list", chainID)
		}
		key := biChainID.String()
		if _, exist := chainIDBlacklistMap[key]; exist {
			return fmt.Errorf("duplicate chain id '%v' in black list", key)
		}
		chainIDBlacklistMap[key] = struct{}{}
	}
	for _, tokenID := range s.TokenIDBlackList {
		if tokenID == "" {
			return errors.New("empty token id in black list")
		}
		key := strings.ToLower(tokenID)
		if _, exist := tokenIDBlacklistMap[key]; exist {
			return fmt.Errorf("duplicate token id '%v' in black list", key)
		}
		tokenIDBlacklistMap[key] = struct{}{}
	}
	log.Info("check server config success")
	return nil
}

// CheckConfig check onchain config storing chain and token configs
func (c *OnchainConfig) CheckConfig() error {
	log.Info("start check onchain config connection")
	if c.Contract == "" {
		return errors.New("onchain must config 'Contract'")
	}
	if len(c.APIAddress) == 0 {
		return errors.New("onchain must config 'APIAddress'")
	}
	callGetChainIDCountData := common.FromHex("0x7b9fb005")
	for _, apiAddress := range c.APIAddress {
		res, err := CallContractWithGateway(apiAddress, c.Contract, callGetChainIDCountData, "latest")
		if err != nil {
			log.Warn("check onchain config connection failed", "gateway", apiAddress, "err", err)
			continue
		}
		chainIDCount, _ := common.GetUint64FromStr(res)
		if chainIDCount == 0 {
			return errors.New("no chain ID in onchain config")
		}
		log.Info("check onchain config connection success", "chainIDCount", chainIDCount)
		return nil
	}
	log.Error("check onchain config connection failed", "gateway", c.APIAddress, "contract", c.Contract)
	return errors.New("check onchain config connection failed")
}
