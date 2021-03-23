package mongodb

import (
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// GetStatusByTokenVerifyError get status by token verify error
func GetStatusByTokenVerifyError(err error) SwapStatus {
	if !tokens.ShouldRegisterSwapForError(err) {
		return TxVerifyFailed
	}
	// TxNotStable status will be reverify at work/verify, add store in result table
	switch err {
	case nil,
		tokens.ErrTxWithWrongMemo,
		tokens.ErrTxWithWrongValue,
		tokens.ErrBindAddrIsContract:
		return TxNotStable
	case tokens.ErrTxSenderNotRegistered:
		return TxSenderNotRegistered
	case tokens.ErrTxWithWrongSender:
		return TxWithWrongSender
	case tokens.ErrTxIncompatible:
		return TxIncompatible
	case tokens.ErrRPCQueryError:
		return RPCQueryError
	default:
		log.Warn("[mongodb] maybe not considered tx verify error", "err", err)
		return TxNotStable
	}
}

// GetRouterSwapStatusByVerifyError get router swap status by verify error
func GetRouterSwapStatusByVerifyError(err error) SwapStatus {
	if !tokens.ShouldRegisterRouterSwapForError(err) {
		return TxVerifyFailed
	}
	switch err {
	case nil:
		return TxNotStable
	case tokens.ErrTxWithWrongValue:
		return TxWithWrongValue
	case tokens.ErrTxWithWrongPath:
		return TxWithWrongPath
	case tokens.ErrMissTokenConfig:
		return MissTokenConfig
	case tokens.ErrNoUnderlyingToken:
		return NoUnderlyingToken
	default:
		return TxVerifyFailed
	}
}
