package router

import (
	"context"
	"encoding/json"
	"math/big"

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

	channels      = make([]chan<- ethtypes.Log, 0, 3)
	subscribes    = make([]ethereum.Subscription, 0, 3)
	updateIDTopic = ethcommon.HexToHash("0x42772a2484b817bd374b06cf7d3ce1e7529d80f9030536688daeb8754e95925f")
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
	SubscribeRouterConfig([]ethcommon.Hash{updateIDTopic})
}

// SubscribeRouterConfig subscribe router config
func SubscribeRouterConfig(topics []ethcommon.Hash) {
	fq := ethereum.FilterQuery{
		Addresses: []ethcommon.Address{routerConfigContract},
		Topics:    [][]ethcommon.Hash{topics},
	}
	for i, cli := range routerConfigClients {
		var ch chan<- ethtypes.Log
		sub, err := cli.SubscribeFilterLogs(routerConfigCtx, fq, ch)
		if err != nil {
			log.Error("subscribe updateID failed", "index", i, "err", err)
			continue
		}
		channels = append(channels, ch)
		subscribes = append(subscribes, sub)
	}
	log.Info("subscribe updateID finished", "subscribes", len(subscribes))
}

// --------------------- getter -----------------------------------

// GetChainConfig abi
func GetChainConfig(chainID *big.Int) (chainCfg *ChainConfig, err error) {
	funcHash := common.FromHex("0x19ed16dc")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	hexData, err := ParseBytesInData(res, 0)
	if err != nil {
		return nil, err
	}
	if len(hexData) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(hexData, &chainCfg)
	if err != nil {
		return nil, err
	}
	return chainCfg, nil
}

func getTokenConfig(funcHash []byte, chainID *big.Int, token string) (tokenCfg *TokenConfig, err error) {
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	copy(data[36:68], common.HexToAddress(token).Hash().Bytes())
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	hexData, err := ParseBytesInData(res, 0)
	if err != nil {
		return nil, err
	}
	if len(hexData) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(hexData, &tokenCfg)
	if err != nil {
		return nil, err
	}
	return tokenCfg, nil
}

// GetTokenConfig abi
func GetTokenConfig(chainID *big.Int, token string) (tokenCfg *TokenConfig, err error) {
	funcHash := common.FromHex("0x6332aec6")
	return getTokenConfig(funcHash, chainID, token)
}

// GetUserTokenConfig abi
func GetUserTokenConfig(chainID *big.Int, token string) (tokenCfg *TokenConfig, err error) {
	funcHash := common.FromHex("0x7ebcc5b1")
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

// GetTokenIDOfToken abi
func GetTokenIDOfToken(chainID *big.Int, token string) (tokenID string, err error) {
	funcHash := common.FromHex("0x0e8257af")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(chainID.Bytes(), 32))
	copy(data[36:68], common.HexToAddress(token).Hash().Bytes())
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

func getOneStrArgData(funcHash []byte, strArg string) []byte {
	length := len(strArg)
	padLength := (length + 31) / 32 * 32
	data := make([]byte, 68+padLength)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte{0x20}, 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(length)).Bytes(), 32))
	copy(data[68:], strArg)
	return data
}

// GetMPCPubkey2 abi
func GetMPCPubkey2(mpcAddress string) (pubkey string, err error) {
	funcHash := common.FromHex("0xffc8fd7b")
	data := getOneStrArgData(funcHash, mpcAddress)
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
	funcHash := common.FromHex("0xaf611ca0")
	data := getOneStrArgData(funcHash, tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return false, err
	}
	return common.GetBigInt(res, 0, 32).Sign() != 0, nil
}

// GetChainIDCount abi
func GetChainIDCount() (count uint64, err error) {
	funcHash := common.FromHex("0x7b9fb005")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return 0, err
	}
	return common.GetBigInt(res, 0, 32).Uint64(), nil
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

// GetChainIDAtIndex abi
func GetChainIDAtIndex(index uint64) (chainID *big.Int, err error) {
	funcHash := common.FromHex("0x0b1bb383")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(index)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	return common.GetBigInt(res, 0, 32), nil
}

// GetChainIDsInRange abi
func GetChainIDsInRange(start, end uint64) (chainIDs []*big.Int, err error) {
	funcHash := common.FromHex("0x60bb8b75")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(start)).Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(end)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	return ParseNumberSliceAsBigIntsInData(res, 0)
}

// GetTokenIDCount abi
func GetTokenIDCount() (count uint64, err error) {
	funcHash := common.FromHex("0x9e1a1087")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return 0, err
	}
	return common.GetBigInt(res, 0, 32).Uint64(), nil
}

// GetAllTokenIDs abi
func GetAllTokenIDs() (tokenIDs []string, err error) {
	funcHash := common.FromHex("0x684a10b3")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return nil, err
	}
	return ParseStringSliceInData(res, 0)
}

// GetTokenIDAtIndex abi
func GetTokenIDAtIndex(index uint64) (tokenID string, err error) {
	funcHash := common.FromHex("0x2915b073")
	data := make([]byte, 36)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(index)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return "", err
	}
	return ParseStringInData(res, 0)
}

// GetTokenIDsInRange abi
func GetTokenIDsInRange(start, end uint64) (tokenIDs []string, err error) {
	funcHash := common.FromHex("0x17394dac")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(start)).Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(end)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	return ParseStringSliceInData(res, 0)
}

// GetMultichainTokenOnChain abi
func GetMultichainTokenOnChain(tokenID string, chainID *big.Int) (tokenAddr string, err error) {
	funcHash := common.FromHex("0xe6729805")
	length := len(tokenID)
	padLength := (length + 31) / 32 * 32
	data := make([]byte, 100+padLength)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte{0x40}, 32))
	copy(data[36:68], common.LeftPadBytes(chainID.Bytes(), 32))
	copy(data[68:100], common.LeftPadBytes(big.NewInt(int64(length)).Bytes(), 32))
	copy(data[100:], tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return "", err
	}
	return common.BigToAddress(common.GetBigInt(res, 0, 32)).String(), nil
}

// GetMultichainTokenCount abi
func GetMultichainTokenCount(tokenID string) (count uint64, err error) {
	funcHash := common.FromHex("0x628180fb")
	data := getOneStrArgData(funcHash, tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return 0, err
	}
	return common.GetBigInt(res, 0, 32).Uint64(), nil
}

// GetMultichainTokensInRange abi
func GetMultichainTokensInRange(tokenID string, start, end uint64) (chainIDs []*big.Int, tokens []string, err error) {
	funcHash := common.FromHex("0x105cc82e")
	length := len(tokenID)
	padLength := (length + 31) / 32 * 32
	data := make([]byte, 132+padLength)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes([]byte{0x60}, 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(start)).Bytes(), 32))
	copy(data[68:100], common.LeftPadBytes(big.NewInt(int64(end)).Bytes(), 32))
	copy(data[100:132], common.LeftPadBytes(big.NewInt(int64(length)).Bytes(), 32))
	copy(data[132:], tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, nil, err
	}
	chainIDs, err = ParseNumberSliceAsBigIntsInData(res, 0)
	if err != nil {
		return nil, nil, err
	}
	tokens, err = ParseAddressSliceInData(res, 32)
	if err != nil {
		return nil, nil, err
	}
	return chainIDs, tokens, nil
}
