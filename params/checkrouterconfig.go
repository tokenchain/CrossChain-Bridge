package params

import (
	"errors"
	"fmt"

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
		if config.MongoDB == nil {
			return errors.New("server must config 'MongoDB'")
		}
		if config.APIServer == nil {
			return errors.New("server must config 'APIServer'")
		}
	}

	if config.Dcrm == nil {
		return errors.New("server must config 'Dcrm'")
	}
	err = config.Dcrm.CheckConfig(isServer)
	if err != nil {
		return err
	}
	log.Info("check dcrm config pass", "isServer", isServer)

	if config.Onchain == nil {
		return errors.New("server must config 'Onchain'")
	}
	err = config.Onchain.CheckConfig()
	if err != nil {
		return err
	}
	log.Info("check onchain config pass")

	return nil
}

// CheckConfig check onchain config
func (c *OnchainConfig) CheckConfig() error {
	callOwnerData := common.FromHex("0x8da5cb5b")
	for _, apiAddress := range c.APIAddress {
		res, err := CallContractWithGateway(apiAddress, c.Contract, callOwnerData, "latest")
		if err != nil {
			log.Warn("check onchain config connection failed", "gateway", apiAddress, "err", err)
			continue
		}
		owner := common.HexToAddress(res)
		if owner == (common.Address{}) {
			continue
		}
		log.Info("check onchain config connection success")
		return nil
	}
	log.Error("wrong onchain config", "gateway", c.APIAddress, "contract", c.Contract)
	return errors.New("check onchain config connection failed")
}
