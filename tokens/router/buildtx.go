package router

import (
	"fmt"
	"math/big"
	"time"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/types"
)

var (
	retryRPCCount    = 3
	retryRPCInterval = 1 * time.Second

	defReserveGasFee = big.NewInt(1e16) // 0.01 ETH

	latestGasPrice *big.Int
)

// BuildRawTransaction build raw tx
func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	if args.Input != nil {
		return nil, fmt.Errorf("forbid build raw swap tx with input data")
	}
	if args.From == "" {
		return nil, fmt.Errorf("forbid empty sender")
	}
	if args.SwapType != tokens.RouterSwapType {
		return nil, tokens.ErrSwapTypeNotSupported
	}

	err = b.buildRouterSwapTxInput(args)
	if err != nil {
		return nil, err
	}

	extra, err := b.setDefaults(args)
	if err != nil {
		return nil, err
	}

	args.Identifier = params.GetIdentifier()

	return b.buildTx(args, extra)
}

func (b *Bridge) buildTx(args *tokens.BuildTxArgs, extra *tokens.EthExtraArgs) (rawTx interface{}, err error) {
	var (
		to       = common.HexToAddress(args.To)
		value    = args.Value
		input    = *args.Input
		nonce    = *extra.Nonce
		gasLimit = *extra.Gas
		gasPrice = extra.GasPrice
	)

	needValue := big.NewInt(0)
	if value != nil && value.Sign() > 0 {
		needValue = value
	}
	needValue = new(big.Int).Add(needValue, defReserveGasFee)
	err = b.checkBalance("", args.From, needValue)
	if err != nil {
		return nil, err
	}

	rawTx = types.NewTransaction(nonce, to, value, gasLimit, gasPrice, input)

	log.Trace("build routerswap raw tx", "swapID", args.SwapID, "swapType", args.SwapType,
		"from", args.From, "to", to.String(), "bind", args.Bind, "nonce", nonce,
		"value", value, "originValue", args.OriginValue, "gasLimit", gasLimit, "gasPrice", gasPrice)

	return rawTx, nil
}

func (b *Bridge) setDefaults(args *tokens.BuildTxArgs) (extra *tokens.EthExtraArgs, err error) {
	if args.Value == nil {
		args.Value = new(big.Int)
	}
	if args.Extra == nil || args.Extra.EthExtra == nil {
		extra = &tokens.EthExtraArgs{}
		args.Extra = &tokens.AllExtras{EthExtra: extra}
	} else {
		extra = args.Extra.EthExtra
	}
	if extra.GasPrice == nil {
		extra.GasPrice, err = b.getGasPrice()
		if err != nil {
			return nil, err
		}
		err = b.adjustSwapGasPrice(args.PairID, extra)
		if err != nil {
			return nil, err
		}
	}
	if extra.Nonce == nil {
		extra.Nonce, err = b.getAccountNonce(args.From)
		if err != nil {
			return nil, err
		}
	}
	if extra.Gas == nil {
		extra.Gas = new(uint64)
		*extra.Gas = b.getDefaultGasLimit(args.PairID)
	}
	return extra, nil
}

func (b *Bridge) getDefaultGasLimit(pairID string) (gasLimit uint64) {
	tokenCfg := b.GetTokenConfig(pairID)
	if tokenCfg != nil {
		gasLimit = tokenCfg.DefaultGasLimit
	}
	if gasLimit == 0 {
		gasLimit = 90000
	}
	return gasLimit
}

func (b *Bridge) getGasPrice() (price *big.Int, err error) {
	for i := 0; i < retryRPCCount; i++ {
		price, err = b.SuggestPrice()
		if err == nil {
			return price, nil
		}
		time.Sleep(retryRPCInterval)
	}
	return nil, err
}

func (b *Bridge) adjustSwapGasPrice(pairID string, extra *tokens.EthExtraArgs) error {
	tokenCfg := b.GetTokenConfig(pairID)
	if tokenCfg == nil {
		return tokens.ErrUnknownPairID
	}
	addPercent := tokenCfg.PlusGasPricePercentage
	if addPercent > 0 {
		extra.GasPrice.Mul(extra.GasPrice, big.NewInt(int64(100+addPercent)))
		extra.GasPrice.Div(extra.GasPrice, big.NewInt(100))
	}
	maxGasPriceFluctPercent := b.ChainConfig.MaxGasPriceFluctPercent
	if maxGasPriceFluctPercent > 0 {
		if latestGasPrice != nil {
			maxFluct := new(big.Int).Set(latestGasPrice)
			maxFluct.Mul(maxFluct, new(big.Int).SetUint64(maxGasPriceFluctPercent))
			maxFluct.Div(maxFluct, big.NewInt(100))
			minGasPrice := new(big.Int).Sub(latestGasPrice, maxFluct)
			if extra.GasPrice.Cmp(minGasPrice) < 0 {
				extra.GasPrice = minGasPrice
			}
		}
		latestGasPrice = extra.GasPrice
	}
	return nil
}

func (b *Bridge) getAccountNonce(from string) (nonceptr *uint64, err error) {
	var nonce uint64
	for i := 0; i < retryRPCCount; i++ {
		nonce, err = b.GetPoolNonce(from, "pending")
		if err == nil {
			break
		}
		time.Sleep(retryRPCInterval)
	}
	if err != nil {
		return nil, err
	}
	nonce = b.AdjustNonce(from, nonce)
	return &nonce, nil
}

func (b *Bridge) checkBalance(token, account string, amount *big.Int) (err error) {
	var balance *big.Int
	for i := 0; i < retryRPCCount; i++ {
		if token != "" {
			balance, err = b.GetErc20Balance(token, account)
		} else {
			balance, err = b.GetBalance(account)
		}
		if err == nil {
			break
		}
		time.Sleep(retryRPCInterval)
	}
	if err == nil && balance.Cmp(amount) < 0 {
		return fmt.Errorf("not enough %v balance. %v < %v", token, balance, amount)
	}
	log.Warn("get balance error", "token", token, "account", account, "err", err)
	return err
}
