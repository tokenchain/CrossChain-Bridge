package router

import (
	"encoding/json"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/common/hexutil"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/rpc/client"
)

// CallOnchainContract call onchain contract
func CallOnchainContract(data hexutil.Bytes, blockNumber string) (result string, err error) {
	onchainCfg := params.GetRouterConfig().Onchain
	reqArgs := map[string]interface{}{
		"to":   onchainCfg.Contract,
		"data": data,
	}
	for _, gateway := range onchainCfg.APIAddress {
		err = client.RPCPost(&result, gateway, "eth_call", reqArgs, blockNumber)
		if err == nil {
			return result, nil
		}
	}
	log.Debug("call onchain contract error", "contract", onchainCfg.Contract, "data", data, "err", err)
	return "", err
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
	hexData, err := ParseBytesInData(common.FromHex(res), 0)
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
	hexData, err := ParseBytesInData(common.FromHex(res), 0)
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
	return ParseStringInData(common.FromHex(res), 0)
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
	return ParseStringInData(common.FromHex(res), 0)
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
	return ParseStringInData(common.FromHex(res), 0)
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
	return ParseStringInData(common.FromHex(res), 0)
}

func getBoolFlagFromStr(str string) (flag bool, err error) {
	bi, err := common.GetBigIntFromStr(str)
	if err != nil {
		return false, err
	}
	return bi.Sign() != 0, nil
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
	return getBoolFlagFromStr(res)
}

// IsTokenIDExist abi
func IsTokenIDExist(tokenID string) (exist bool, err error) {
	funcHash := common.FromHex("0xaf611ca0")
	data := getOneStrArgData(funcHash, tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return false, err
	}
	return getBoolFlagFromStr(res)
}

// GetChainIDCount abi
func GetChainIDCount() (count uint64, err error) {
	funcHash := common.FromHex("0x7b9fb005")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return 0, err
	}
	return common.GetUint64FromStr(res)
}

// GetAllChainIDs abi
func GetAllChainIDs() (chainIDs []*big.Int, err error) {
	funcHash := common.FromHex("0xe27112d5")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return nil, err
	}
	return ParseNumberSliceAsBigIntsInData(common.FromHex(res), 0)
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
	return common.GetBigIntFromStr(res)
}

// GetChainIDInRange abi
func GetChainIDInRange(start, end uint64) (chainIDs []*big.Int, err error) {
	funcHash := common.FromHex("0x454a8f28")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(start)).Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(end)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	return ParseNumberSliceAsBigIntsInData(common.FromHex(res), 0)
}

// GetTokenIDCount abi
func GetTokenIDCount() (count uint64, err error) {
	funcHash := common.FromHex("0x9e1a1087")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return 0, err
	}
	return common.GetUint64FromStr(res)
}

// GetAllTokenIDs abi
func GetAllTokenIDs() (tokenIDs []string, err error) {
	funcHash := common.FromHex("0x684a10b3")
	res, err := CallOnchainContract(funcHash, "latest")
	if err != nil {
		return nil, err
	}
	return ParseStringSliceInData(common.FromHex(res), 0)
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
	return ParseStringInData(common.FromHex(res), 0)
}

// GetTokenIDInRange abi
func GetTokenIDInRange(start, end uint64) (tokenIDs []string, err error) {
	funcHash := common.FromHex("0xb22cb69a")
	data := make([]byte, 68)
	copy(data[:4], funcHash)
	copy(data[4:36], common.LeftPadBytes(big.NewInt(int64(start)).Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(big.NewInt(int64(end)).Bytes(), 32))
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return nil, err
	}
	return ParseStringSliceInData(common.FromHex(res), 0)
}

// GetPeerTokenOnChain abi
func GetPeerTokenOnChain(tokenID string, chainID *big.Int) (tokenAddr string, err error) {
	funcHash := common.FromHex("0x74268f22")
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
	return common.HexToAddress(res).String(), nil
}

// GetPeerTokenCount abi
func GetPeerTokenCount(tokenID string) (count uint64, err error) {
	funcHash := common.FromHex("0x7f5a6c71")
	data := getOneStrArgData(funcHash, tokenID)
	res, err := CallOnchainContract(data, "latest")
	if err != nil {
		return 0, err
	}
	bi, err := common.GetBigIntFromStr(res)
	if err != nil {
		return 0, err
	}
	return bi.Uint64(), nil
}

// GetPeerTokenInRange abi
func GetPeerTokenInRange(tokenID string, start, end uint64) (chainIDs []*big.Int, tokens []string, err error) {
	funcHash := common.FromHex("0xeb1c5c87")
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
	chainIDs, err = ParseNumberSliceAsBigIntsInData(common.FromHex(res), 0)
	if err != nil {
		return nil, nil, err
	}
	tokens, err = ParseAddressSliceInData(common.FromHex(res), 32)
	if err != nil {
		return nil, nil, err
	}
	return chainIDs, tokens, nil
}
