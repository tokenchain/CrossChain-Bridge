package eth

import (
	"errors"
	"math/big"
	"time"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// router contract's func hashs
var (
	defSwapDeadlineOffset = int64(36000)

	// anySwapIn(bytes32 txs, address token, address to, uint amount, uint fromChainID)
	AnySwapInFuncHash = common.FromHex("0x825bb13c")
	// anySwapInUnderlying(bytes32 txs, address token, address to, uint amount, uint fromChainID)
	AnySwapInUnderlyingFuncHash = common.FromHex("0x3f88de89")
	// anySwapInExactTokensForTokens(bytes32 txs, uint amountIn, uint amountOutMin, address[] path, address to, uint deadline, uint fromChainID)
	AnySwapInExactTokensForTokensFuncHash = common.FromHex("0x2fc1e728")
	// anySwapInExactTokensForNative(bytes32 txs, uint amountIn, uint amountOutMin, address[] path, address to, uint deadline, uint fromChainID)
	AnySwapInExactTokensForNativeFuncHash = common.FromHex("0x52a397d5")
)

func (b *Bridge) buildSwapoutTxInput(args *tokens.BuildTxArgs, tokenCfg *tokens.TokenConfig) (err error) {
	isRouterSwap := params.IsRouterSwap() || b.ChainConfig.RouterContract != ""
	if !isRouterSwap && !b.IsSrc {
		return tokens.ErrBuildSwapTxInWrongEndpoint
	}
	switch {
	case isRouterSwap:
		if len(args.Path) > 0 && args.AmountOutMin != nil {
			return b.buildRouterSwapoutTradeTxInput(args)
		}
		return b.buildRouterSwapoutTxInput(args)
	case tokenCfg.IsErc20():
		return b.buildErc20SwapoutTxInput(args)
	default:
		input := []byte(tokens.UnlockMemoPrefix + args.SwapID)
		args.Input = &input // input
		args.To = args.Bind // to
		return nil
	}
}

func (b *Bridge) buildErc20SwapoutTxInput(args *tokens.BuildTxArgs) (err error) {
	token, receiver, amount, err := b.checkSwapoutReceiverAndAmount(args)
	if err != nil {
		return err
	}

	funcHash := erc20CodeParts["transfer"]
	input := PackDataWithFuncHash(funcHash, receiver, amount)
	args.Input = &input             // input
	args.To = token.ContractAddress // to

	return b.checkBalance(token.ContractAddress, token.DcrmAddress, amount)
}

func (b *Bridge) buildRouterSwapoutTxInput(args *tokens.BuildTxArgs) (err error) {
	token, receiver, amount, err := b.checkSwapoutReceiverAndAmount(args)
	if err != nil {
		return err
	}

	var funcHash []byte
	if args.ForUnderlying {
		funcHash = AnySwapInUnderlyingFuncHash
	} else {
		funcHash = AnySwapInFuncHash
	}

	input := PackDataWithFuncHash(funcHash, args.Token, receiver, amount, args.FromChainID)
	args.Input = &input                    // input
	args.To = b.ChainConfig.RouterContract // to

	return b.checkBalance(token.ContractAddress, token.DcrmAddress, amount)
}

func (b *Bridge) buildRouterSwapoutTradeTxInput(args *tokens.BuildTxArgs) (err error) {
	token, receiver, amount, err := b.checkSwapoutReceiverAndAmount(args)
	if err != nil {
		return err
	}

	var funcHash []byte
	if args.ForNative {
		funcHash = AnySwapInExactTokensForNativeFuncHash
	} else {
		funcHash = AnySwapInExactTokensForTokensFuncHash
	}

	swapDeadlineOffset := b.ChainConfig.SwapDeadlineOffset
	if swapDeadlineOffset == 0 {
		swapDeadlineOffset = defSwapDeadlineOffset
	}
	deadline := time.Now().Unix() + swapDeadlineOffset

	input := PackDataWithFuncHash(funcHash, args.SwapID, amount, args.AmountOutMin, toAddresses(args.Path), receiver, deadline, args.FromChainID)
	args.Input = &input                    // input
	args.To = b.ChainConfig.RouterContract // to

	return b.checkBalance(token.ContractAddress, token.DcrmAddress, amount)
}

func (b *Bridge) checkSwapoutReceiverAndAmount(args *tokens.BuildTxArgs) (
	token *tokens.TokenConfig, receiver common.Address, amount *big.Int, err error) {
	pairID := args.PairID
	token = b.GetTokenConfig(pairID)
	if token == nil {
		return token, receiver, amount, tokens.ErrUnknownPairID
	}
	receiver = common.HexToAddress(args.Bind)
	if receiver == (common.Address{}) || !common.IsHexAddress(args.Bind) {
		log.Warn("swapout to wrong receiver", "receiver", args.Bind)
		return token, receiver, amount, errors.New("can not swapout to empty or invalid receiver")
	}
	amount = tokens.CalcSwappedValue(pairID, args.OriginValue, false)
	return token, receiver, amount, nil
}

func toAddresses(path []string) []common.Address {
	addresses := make([]common.Address, len(path))
	for i, addr := range path {
		addresses[i] = common.HexToAddress(addr)
	}
	return addresses
}
