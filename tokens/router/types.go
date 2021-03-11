package router

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/tools/crypto"
)

// ChainConfig struct
type ChainConfig struct {
	BlockChain              string
	ChainID                 string
	Confirmations           uint64
	InitialHeight           uint64
	WaitTimeToReplace       int64 // seconds
	MaxReplaceCount         int
	RouterContract          string
	SwapDeadlineOffset      int64  `json:",omitempty"` // seconds
	PlusGasPricePercentage  uint64 `json:",omitempty"`
	MaxGasPriceFluctPercent uint64 `json:",omitempty"`

	// cached value
	routerMPC       string
	routerMPCPubkey string
}

// TokenConfig struct
type TokenConfig struct {
	ID                string
	Name              string
	Symbol            string
	Decimals          uint8
	ContractAddress   string
	MaximumSwap       float64 // whole unit (eg. BTC, ETH, FSN), not Satoshi
	MinimumSwap       float64 // whole unit
	BigValueThreshold float64
	SwapFeeRate       float64
	MaximumSwapFee    float64
	MinimumSwapFee    float64
	DefaultGasLimit   uint64

	// calced value
	maxSwap          *big.Int
	minSwap          *big.Int
	maxSwapFee       *big.Int
	minSwapFee       *big.Int
	bigValThreshhold *big.Int
}

// GatewayConfig struct
type GatewayConfig struct {
	APIAddress []string
}

// CheckConfig check chain config
func (c *ChainConfig) CheckConfig() error {
	if c.BlockChain == "" {
		return errors.New("chain must config 'BlockChain'")
	}
	if c.ChainID == "" {
		return errors.New("chain must config 'ChainID'")
	}
	if _, ok := new(big.Int).SetString(c.ChainID, 0); !ok {
		return errors.New("chain with wrong 'ChainID'")
	}
	if c.Confirmations == 0 {
		return errors.New("chain must config nonzero 'Confirmations'")
	}
	if c.RouterContract == "" {
		return errors.New("chain must config 'RouterContract'")
	}
	maxPlusGasPricePercentage := uint64(10000)
	if c.PlusGasPricePercentage > maxPlusGasPricePercentage {
		return errors.New("too large 'PlusGasPricePercentage' value")
	}
	return nil
}

// GetRouterMPC get router mpc
func (c *ChainConfig) GetRouterMPC() string {
	return c.routerMPC
}

// GetRouterMPCPubkey get router mpc public key
func (c *ChainConfig) GetRouterMPCPubkey() string {
	return c.routerMPCPubkey
}

// CheckConfig check token config
//nolint:gocyclo // keep TokenConfig check as whole
func (c *TokenConfig) CheckConfig(isSrc bool) error {
	if c.ID == "" {
		return errors.New("token must config 'ID'")
	}
	if c.ContractAddress == "" {
		return errors.New("token must config 'ContractAddress'")
	}
	if c.MaximumSwap < 0 {
		return errors.New("token must config 'MaximumSwap' (non-negative)")
	}
	if c.MinimumSwap < 0 {
		return errors.New("token must config 'MinimumSwap' (non-negative)")
	}
	if c.MinimumSwap > c.MaximumSwap {
		return errors.New("wrong token config, MinimumSwap > MaximumSwap")
	}
	if c.SwapFeeRate < 0 || c.SwapFeeRate > 1 {
		return errors.New("token must config 'SwapFeeRate' (in range (0,1))")
	}
	if c.MaximumSwapFee < 0 {
		return errors.New("token must config 'MaximumSwapFee' (non-negative)")
	}
	if c.MinimumSwapFee < 0 {
		return errors.New("token must config 'MinimumSwapFee' (non-negative)")
	}
	if c.MinimumSwapFee > c.MaximumSwapFee {
		return errors.New("wrong token config, MinimumSwapFee > MaximumSwapFee")
	}
	if c.MinimumSwap < c.MinimumSwapFee {
		return errors.New("wrong token config, MinimumSwap < MinimumSwapFee")
	}
	if c.SwapFeeRate == 0.0 && c.MinimumSwapFee > 0.0 {
		return errors.New("wrong token config, MinimumSwapFee should be 0 if SwapFeeRate is 0")
	}
	if c.BigValueThreshold == 0 {
		return errors.New("token must config 'BigValueThreshold' (non-negative)")
	}

	c.calcAndStoreValue()
	return nil
}

// GetBigValueThreshold get big vaule threshold
func (c *TokenConfig) GetBigValueThreshold() *big.Int {
	return c.bigValThreshhold
}

func (c *TokenConfig) calcAndStoreValue() {
	c.maxSwap = toBits(c.MaximumSwap, c.Decimals)
	c.minSwap = toBits(c.MinimumSwap, c.Decimals)
	c.maxSwapFee = toBits(c.MaximumSwapFee, c.Decimals)
	c.minSwapFee = toBits(c.MinimumSwapFee, c.Decimals)
	c.bigValThreshhold = toBits(c.BigValueThreshold+0.0001, c.Decimals)
}

// VerifyDcrmPublicKey verify dcrm address and public key is matching
func VerifyDcrmPublicKey(dcrmAddress, dcrmPubkey string) error {
	if !common.IsHexAddress(dcrmAddress) {
		return fmt.Errorf("wrong dcrm address '%v'", dcrmAddress)
	}
	pkBytes := common.FromHex(dcrmPubkey)
	if len(pkBytes) != 65 || pkBytes[0] != 4 {
		return fmt.Errorf("wrong dcrm public key '%v'", dcrmPubkey)
	}
	pubKey := ecdsa.PublicKey{
		Curve: crypto.S256(),
		X:     new(big.Int).SetBytes(pkBytes[1:33]),
		Y:     new(big.Int).SetBytes(pkBytes[33:65]),
	}
	pubAddr := crypto.PubkeyToAddress(pubKey)
	if !strings.EqualFold(pubAddr.String(), dcrmAddress) {
		return fmt.Errorf("dcrm address %v and public key address %v is not match", dcrmAddress, pubAddr.String())
	}
	return nil
}
