package router

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/common/hexutil"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	ethereum "github.com/fsn-dev/fsn-go-sdk/efsn"
	ethcommon "github.com/fsn-dev/fsn-go-sdk/efsn/common"
	ethtypes "github.com/fsn-dev/fsn-go-sdk/efsn/core/types"
	"github.com/fsn-dev/fsn-go-sdk/efsn/ethclient"
)

var (
	routerConfigContract ethcommon.Address
	routerConfigClients  []*ethclient.Client
	routerConfigCtx      = context.Background()

	channels   = make([]chan ethtypes.Log, 0, 3)
	subscribes = make([]ethereum.Subscription, 0, 3)

	// topic of event 'UpdateConfig()'
	updateConfigTopic = ethcommon.HexToHash("0x22590461e7ba17e1fe7580cb0ea47f283d3b2248f04873dfbe926d08fe4c5ab9")

	latestUpdateIDBlock uint64
)

// InitRouterConfigClients init router config clients
func InitRouterConfigClients() {
	var err error
	onchainCfg := params.GetRouterConfig().Onchain
	routerConfigContract = ethcommon.HexToAddress(onchainCfg.Contract)
	routerConfigClients = make([]*ethclient.Client, len(onchainCfg.APIAddress))
	for i, gateway := range onchainCfg.APIAddress {
		routerConfigClients[i], err = ethclient.Dial(gateway)
		if err != nil {
			log.Fatal("init router config clients failed", "gateway", gateway, "err", err)
		}
	}
}

// CallOnchainContract call onchain contract
func CallOnchainContract(data hexutil.Bytes, blockNumber string) (result []byte, err error) {
	msg := ethereum.CallMsg{
		To:   &routerConfigContract,
		Data: data,
	}
	for _, cli := range routerConfigClients {
		result, err = cli.CallContract(routerConfigCtx, msg, nil)
		if err == nil {
			return result, nil
		}
	}
	log.Debug("call onchain contract error", "contract", routerConfigContract.String(), "data", data, "err", err)
	return nil, err
}

// SubscribeUpdateID subscribe update ID and reload configs
func SubscribeUpdateID() {
	SubscribeRouterConfig([]ethcommon.Hash{updateConfigTopic})
	for _, ch := range channels {
		go processUpdateID(ch)
	}
}

func processUpdateID(ch <-chan ethtypes.Log) {
	for {
		rlog := <-ch

		// sleep random in a second to mess steps
		rNum, _ := rand.Int(rand.Reader, big.NewInt(1000))
		time.Sleep(time.Duration(rNum.Uint64()) * time.Millisecond)

		blockNumber := rlog.BlockNumber
		oldBlock := atomic.LoadUint64(&latestUpdateIDBlock)
		if blockNumber > oldBlock {
			atomic.StoreUint64(&latestUpdateIDBlock, blockNumber)
			ReloadRouterConfig()
		}
	}
}

// SubscribeRouterConfig subscribe router config
func SubscribeRouterConfig(topics []ethcommon.Hash) {
	fq := ethereum.FilterQuery{
		Addresses: []ethcommon.Address{routerConfigContract},
		Topics:    [][]ethcommon.Hash{topics},
	}
	for i, cli := range routerConfigClients {
		ch := make(chan ethtypes.Log)
		sub, err := cli.SubscribeFilterLogs(routerConfigCtx, fq, ch)
		if err != nil {
			log.Error("subscribe 'UpdateConfig' event failed", "index", i, "err", err)
			continue
		}
		channels = append(channels, ch)
		subscribes = append(subscribes, sub)
	}
	log.Info("subscribe 'UpdateConfig' event finished", "subscribes", len(subscribes))
}

func parseChainConfig(data []byte) (config *ChainConfig, err error) {
	offset, overflow := common.GetUint64(data, 0, 32)
	if overflow {
		return nil, errParseDataError
	}
	if uint64(len(data)) < offset+12*32 {
		return nil, errParseDataError
	}
	data = data[32:]
	config = &ChainConfig{}
	config.BlockChain, err = ParseStringInData(data, 0)
	if err != nil {
		return nil, errParseDataError
	}
	config.RouterContract = common.BytesToAddress(common.GetData(data, 32, 32)).String()
	config.Confirmations = common.GetBigInt(data, 64, 32).Uint64()
	config.InitialHeight = common.GetBigInt(data, 96, 32).Uint64()
	config.WaitTimeToReplace = common.GetBigInt(data, 128, 32).Int64()
	config.MaxReplaceCount = int(common.GetBigInt(data, 160, 32).Int64())
	config.SwapDeadlineOffset = common.GetBigInt(data, 192, 32).Int64()
	config.PlusGasPricePercentage = common.GetBigInt(data, 224, 32).Uint64()
	config.MaxGasPriceFluctPercent = common.GetBigInt(data, 256, 32).Uint64()
	config.DefaultGasLimit = common.GetBigInt(data, 288, 32).Uint64()
	return config, nil
}

// GetChainConfig abi
func GetChainConfig(chainID *big.Int) (*ChainConfig, error) {
	funcHash := common.FromHex("0x19ed16dc")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	config, err := parseChainConfig(res)
	if err != nil {
		return nil, err
	}
	config.ChainID = chainID.String()
	return config, nil
}

func parseTokenConfig(data []byte) (config *TokenConfig, err error) {
	if uint64(len(data)) < 9*32 {
		return nil, errParseDataError
	}
	decimals := uint8(common.GetBigInt(data, 0, 32).Uint64())
	contractAddress := common.BytesToAddress(common.GetData(data, 32, 32)).String()
	contractVersion := common.GetBigInt(data, 64, 32).Uint64()
	maximumSwap := common.GetBigInt(data, 96, 32)
	minimumSwap := common.GetBigInt(data, 128, 32)
	bigValueThreshold := common.GetBigInt(data, 160, 32)
	swapFeeRate := common.GetBigInt(data, 192, 32)
	maximumSwapFee := common.GetBigInt(data, 224, 32)
	minimumSwapFee := common.GetBigInt(data, 256, 32)
	config = &TokenConfig{
		Decimals:          decimals,
		ContractAddress:   contractAddress,
		ContractVersion:   contractVersion,
		MaximumSwap:       FromBits(maximumSwap, decimals),
		MinimumSwap:       FromBits(minimumSwap, decimals),
		BigValueThreshold: FromBits(bigValueThreshold, decimals),
		SwapFeeRate:       FromBits(swapFeeRate, 6),
		MaximumSwapFee:    FromBits(maximumSwapFee, decimals),
		MinimumSwapFee:    FromBits(minimumSwapFee, decimals),
	}
	return config, err
}

func getTokenConfig(funcHash []byte, chainID *big.Int, token string) (*TokenConfig, error) {
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte(token), 32))
	copy(data[36:68], common.LeftPadBytes(chainID.Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	config, err := parseTokenConfig(res)
	if err != nil {
		return nil, err
	}
	config.TokenID = token
	return config, nil
}

// GetTokenConfig abi
func GetTokenConfig(chainID *big.Int, token string) (tokenCfg *TokenConfig, err error) {
	funcHash := common.FromHex("0xa5bc5953")
	return getTokenConfig(funcHash, chainID, token)
}

// GetUserTokenConfig abi
func GetUserTokenConfig(chainID *big.Int, token string) (tokenCfg *TokenConfig, err error) {
	funcHash := common.FromHex("0xb329b08a")
	return getTokenConfig(funcHash, chainID, token)
}

// GetCustomConfig abi
func GetCustomConfig(chainID *big.Int, key string) (string, error) {
	funcHash := common.FromHex("0x61387d61")
	length := len(key)
	padLength := (length + 31) / 32 * 32
	data := make([]byte, 100+padLength)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes([]byte{0x40}, 32))
	copy(data[68:100], common.LeftPadBytes(big.NewInt(int64(length)).Bytes(), 32))
	copy(data[100:], key)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return "", err
	}
	return ParseStringInData(res, 0)
}

// GetMPCPubkey abi
func GetMPCPubkey(mpcAddress string) (pubkey string, err error) {
	funcHash := common.FromHex("0x58bb97fb")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.HexToAddress(mpcAddress).Hash().Bytes())
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return "", err
	}
	return ParseStringInData(res, 0)
}

// IsChainIDExist abi
func IsChainIDExist(chainID *big.Int) (exist bool, err error) {
	funcHash := common.FromHex("0xfd15ea70")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return false, err
	}
	return common.GetBigInt(res, 0, 32).Sign() != 0, nil
}

// IsTokenIDExist abi
func IsTokenIDExist(tokenID string) (exist bool, err error) {
	funcHash := common.FromHex("0x97c9877f")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte(tokenID), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return false, err
	}
	return common.GetBigInt(res, 0, 32).Sign() != 0, nil
}

// GetAllChainIDs abi
func GetAllChainIDs() (chainIDs []*big.Int, err error) {
	funcHash := common.FromHex("0xe27112d5")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return nil, err
	}
	return ParseNumberSliceAsBigIntsInData(res, 0)
}

// GetAllTokenIDs abi
func GetAllTokenIDs() (tokenIDs []string, err error) {
	funcHash := common.FromHex("0x684a10b3")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return nil, err
	}
	bns, err := ParseNumberSliceAsBigIntsInData(res, 0)
	if err != nil {
		return nil, err
	}
	tokenIDs = make([]string, len(bns))
	for i, bn := range bns {
		tokenIDs[i] = string(bn.Bytes())
	}
	return tokenIDs, nil
}

// GetMultichainToken abi
func GetMultichainToken(tokenID string, chainID *big.Int) (tokenAddr string, err error) {
	funcHash := common.FromHex("0xec85d336")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte(tokenID), 32))
	copy(data[36:68], common.LeftPadBytes(chainID.Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return "", err
	}
	return common.BigToAddress(common.GetBigInt(res, 0, 32)).String(), nil
}

func parseMultichainTokens(data []byte) (tokenAddrs []string, chainIDs []*big.Int, err error) {
	offset, overflow := common.GetUint64(data, 0, 32)
	if overflow {
		return nil, nil, errParseDataError
	}
	length, overflow := common.GetUint64(data, offset, 32)
	if overflow {
		return nil, nil, errParseDataError
	}
	if uint64(len(data)) < offset+32+length*64 {
		return nil, nil, errParseDataError
	}
	tokenAddrs = make([]string, length)
	chainIDs = make([]*big.Int, length)
	data = data[offset+32:]
	for i := uint64(0); i < length; i++ {
		chainIDs[i] = common.GetBigInt(data, i*64, 32)
		tokenAddrs[i] = common.BytesToAddress(common.GetData(data, i*64+32, 32)).String()
	}
	return tokenAddrs, chainIDs, nil
}

// GetAllMultichainTokens abi
func GetAllMultichainTokens(tokenID string) (tokenAddrs []string, chainIDs []*big.Int, err error) {
	funcHash := common.FromHex("231c77be")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte(tokenID), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, nil, err
	}
	return parseMultichainTokens(res)
}
