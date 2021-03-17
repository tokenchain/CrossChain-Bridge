package router

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/common/hexutil"
)

// token types (should be all upper case)
const (
	ERC20TokenType = "ERC20"
)

var erc20CodeParts = map[string][]byte{
	"name":         common.FromHex("0x06fdde03"),
	"symbol":       common.FromHex("0x95d89b41"),
	"decimals":     common.FromHex("0x313ce567"),
	"totalSupply":  common.FromHex("0x18160ddd"),
	"balanceOf":    common.FromHex("0x70a08231"),
	"transfer":     common.FromHex("0xa9059cbb"),
	"transferFrom": common.FromHex("0x23b872dd"),
	"approve":      common.FromHex("0x095ea7b3"),
	"allowance":    common.FromHex("0xdd62ed3e"),
	"LogTransfer":  common.FromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
	"LogApproval":  common.FromHex("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
}

// GetErc20TotalSupply get erc20 total supply of address
func (b *Bridge) GetErc20TotalSupply(contract string) (*big.Int, error) {
	data := make(hexutil.Bytes, 4)
	copy(data[:4], erc20CodeParts["totalSupply"])
	result, err := b.CallContract(contract, data, "latest")
	if err != nil {
		return nil, err
	}
	return common.GetBigIntFromStr(result)
}

// GetErc20Balance get erc20 balacne of address
func (b *Bridge) GetErc20Balance(contract, address string) (*big.Int, error) {
	data := make(hexutil.Bytes, 36)
	copy(data[:4], erc20CodeParts["balanceOf"])
	copy(data[4:], common.HexToAddress(address).Hash().Bytes())
	result, err := b.CallContract(contract, data, "latest")
	if err != nil {
		return nil, err
	}
	return common.GetBigIntFromStr(result)
}

// GetErc20Decimals get erc20 decimals
func (b *Bridge) GetErc20Decimals(contract string) (uint8, error) {
	data := make(hexutil.Bytes, 4)
	copy(data[:4], erc20CodeParts["decimals"])
	result, err := b.CallContract(contract, data, "latest")
	if err != nil {
		return 0, err
	}
	decimals, err := common.GetUint64FromStr(result)
	return uint8(decimals), err
}

// GetTokenBalance api
func (b *Bridge) GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error) {
	switch strings.ToUpper(tokenType) {
	case ERC20TokenType:
		return b.GetErc20Balance(tokenAddress, accountAddress)
	default:
		return nil, fmt.Errorf("[%v] can not get token balance of token with type '%v'", b.ChainConfig.BlockChain, tokenType)
	}
}

// GetTokenSupply impl
func (b *Bridge) GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error) {
	switch strings.ToUpper(tokenType) {
	case ERC20TokenType:
		return b.GetErc20TotalSupply(tokenAddress)
	default:
		return nil, fmt.Errorf("[%v] can not get token supply of token with type '%v'", b.ChainConfig.BlockChain, tokenType)
	}
}

// GetRouterMPC get router contract's mpc address
func (b *Bridge) GetRouterMPC(routerContract string) (string, error) {
	data := common.FromHex("0xf75c2664")
	res, err := b.CallContract(routerContract, data, "latest")
	if err != nil {
		return "", err
	}
	return ParseStringInData(common.FromHex(res), 0)
}

// GetVaultAddress get token's vault (router) address
func (b *Bridge) GetVaultAddress(tokenAddr string) (string, error) {
	data := common.FromHex("0xfbfa77cf")
	res, err := b.CallContract(tokenAddr, data, "latest")
	if err != nil {
		return "", err
	}
	return ParseStringInData(common.FromHex(res), 0)
}

func parseSliceInData(data []byte, pos uint64) (offset, length uint64, err error) {
	offset, overflow := common.GetUint64(data, pos, 32)
	if overflow {
		return 0, 0, errParseDataError
	}
	length, overflow = common.GetUint64(data, offset, 32)
	if overflow {
		return 0, 0, errParseDataError
	}
	offset += 32
	if uint64(len(data)) < offset+length*32 {
		return 0, 0, errParseDataError
	}
	return offset, length, nil
}

// ParseAddressSliceInData parse
func ParseAddressSliceInData(data []byte, pos uint64) ([]string, error) {
	offset, length, err := parseSliceInData(data, pos)
	if err != nil {
		return nil, err
	}
	path := make([]string, length)
	for i := uint64(0); i < length; i++ {
		path[i] = common.BytesToAddress(common.GetData(data, offset, 32)).String()
		offset += 32
	}
	return path, nil
}

// ParseAddressSliceAsAddressesInData parse
func ParseAddressSliceAsAddressesInData(data []byte, pos uint64) ([]common.Address, error) {
	offset, length, err := parseSliceInData(data, pos)
	if err != nil {
		return nil, err
	}
	path := make([]common.Address, length)
	for i := uint64(0); i < length; i++ {
		path[i] = common.BytesToAddress(common.GetData(data, offset, 32))
		offset += 32
	}
	return path, nil
}

// ParseNumberSliceInData parse
func ParseNumberSliceInData(data []byte, pos uint64) ([]string, error) {
	offset, length, err := parseSliceInData(data, pos)
	if err != nil {
		return nil, err
	}
	results := make([]string, length)
	for i := uint64(0); i < length; i++ {
		results[i] = common.GetBigInt(data, offset, 32).String()
		offset += 32
	}
	return results, nil
}

// ParseNumberSliceAsBigIntsInData parse
func ParseNumberSliceAsBigIntsInData(data []byte, pos uint64) ([]*big.Int, error) {
	offset, length, err := parseSliceInData(data, pos)
	if err != nil {
		return nil, err
	}
	results := make([]*big.Int, length)
	for i := uint64(0); i < length; i++ {
		results[i] = common.GetBigInt(data, offset, 32)
		offset += 32
	}
	return results, nil
}

// ParseStringSliceInData parse
func ParseStringSliceInData(data []byte, pos uint64) ([]string, error) {
	offset, length, err := parseSliceInData(data, pos)
	if err != nil {
		return nil, err
	}
	// new data for inner array
	data = data[offset:]
	offset = 0
	results := make([]string, length)
	for i := uint64(0); i < length; i++ {
		str, err := ParseStringInData(data, offset)
		if err != nil {
			return nil, err
		}
		results[i] = str
		offset += 32
	}
	return results, nil
}

// ParseStringInData parse
func ParseStringInData(data []byte, pos uint64) (string, error) {
	offset, overflow := common.GetUint64(data, pos, 32)
	if overflow {
		return "", errParseDataError
	}
	length, overflow := common.GetUint64(data, offset, 32)
	if overflow {
		return "", errParseDataError
	}
	if uint64(len(data)) < offset+32+length {
		return "", errParseDataError
	}
	return string(common.GetData(data, offset+32, length)), nil
}

// ParseBytesInData parse
func ParseBytesInData(data []byte, pos uint64) ([]byte, error) {
	offset, overflow := common.GetUint64(data, pos, 32)
	if overflow {
		return nil, errParseDataError
	}
	length, overflow := common.GetUint64(data, offset, 32)
	if overflow {
		return nil, errParseDataError
	}
	if uint64(len(data)) < offset+32+length {
		return nil, errParseDataError
	}
	return common.GetData(data, offset+32, length), nil
}
