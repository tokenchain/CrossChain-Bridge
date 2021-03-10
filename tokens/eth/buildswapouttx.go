package eth

import (
	"errors"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) buildSwapoutTxInput(args *tokens.BuildTxArgs, tokenCfg *tokens.TokenConfig) (err error) {
	if !b.IsSrc {
		return tokens.ErrBuildSwapTxInWrongEndpoint
	}
	switch {
	case tokenCfg.IsErc20():
		return b.buildErc20SwapoutTxInput(args)
	default:
		input := []byte(tokens.UnlockMemoPrefix + args.SwapID)
		args.Input = &input // input
		args.To = args.Bind // to
		args.Value = tokens.CalcSwapValue(tokenCfg, args.OriginValue)
	}
	return nil
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
