package router

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

func (b *Bridge) buildRouterSwapTxInput(args *tokens.BuildTxArgs) (err error) {
	if !params.IsRouterSwap() || b.ChainConfig.RouterContract == "" {
		return tokens.ErrRouterSwapNotSupport
	}
	if len(args.Path) > 0 && args.AmountOutMin != nil {
		return b.buildRouterSwapTradeTxInput(args)
	}
	return b.buildRouterSwapoutTxInput(args)
}

func (b *Bridge) buildRouterSwapoutTxInput(args *tokens.BuildTxArgs) (err error) {
	token, receiver, amount, err := b.getReceiverAndAmount(args)
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
	args.SwapValue = amount                // swapValue

	return b.checkBalance(token.ContractAddress, token.DcrmAddress, amount)
}

func (b *Bridge) buildRouterSwapTradeTxInput(args *tokens.BuildTxArgs) (err error) {
	token, receiver, amount, err := b.getReceiverAndAmount(args)
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
	args.SwapValue = amount                // swapValue

	return b.checkBalance(token.ContractAddress, token.DcrmAddress, amount)
}

func (b *Bridge) getReceiverAndAmount(args *tokens.BuildTxArgs) (tokenCfg *tokens.TokenConfig, receiver common.Address, amount *big.Int, err error) {
	tokenCfg = b.GetTokenConfig(args.Path[0])
	if tokenCfg == nil {
		return tokenCfg, receiver, amount, tokens.ErrMissTokenConfig
	}
	receiver = common.HexToAddress(args.Bind)
	if receiver == (common.Address{}) || !common.IsHexAddress(args.Bind) {
		log.Warn("swapout to wrong receiver", "receiver", args.Bind)
		return tokenCfg, receiver, amount, errors.New("can not swapout to empty or invalid receiver")
	}
	fromBridge := GetBridgeByChainID(args.FromChainID.String())
	amount = fromBridge.CalcSwapValue(args.Token, args.OriginValue)
	return tokenCfg, receiver, amount, nil
}

func toAddresses(path []string) []common.Address {
	addresses := make([]common.Address, len(path))
	for i, addr := range path {
		addresses[i] = common.HexToAddress(addr)
	}
	return addresses
}
