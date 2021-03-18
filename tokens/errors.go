package tokens

import (
	"errors"
)

// common errors
var (
	ErrSwapTypeNotSupported          = errors.New("swap type not supported in this endpoint")
	ErrBridgeSourceNotSupported      = errors.New("bridge source not supported")
	ErrBridgeDestinationNotSupported = errors.New("bridge destination not supported")
	ErrUnknownSwapType               = errors.New("unknown swap type")
	ErrMsgHashMismatch               = errors.New("message hash mismatch")
	ErrWrongCountOfMsgHashes         = errors.New("wrong count of msg hashed")
	ErrWrongRawTx                    = errors.New("wrong raw tx")
	ErrWrongExtraArgs                = errors.New("wrong extra args")
	ErrNoBtcBridge                   = errors.New("no btc bridge exist")
	ErrWrongSwapinTxType             = errors.New("wrong swapin tx type")
	ErrBuildSwapTxInWrongEndpoint    = errors.New("build swap in/out tx in wrong endpoint")
	ErrTxBeforeInitialHeight         = errors.New("transaction before initial block height")
	ErrAddressIsInBlacklist          = errors.New("address is in black list")
	ErrSwapInBlacklist               = errors.New("swap is in black list")
	ErrRouterSwapNotSupport          = errors.New("bridge does not support router swap")
	ErrNoBridgeForChainID            = errors.New("no bridge for chain id")

	ErrTxNotFound           = errors.New("tx not found")
	ErrTxNotStable          = errors.New("tx not stable")
	ErrTxWithWrongReceiver  = errors.New("tx with wrong receiver")
	ErrTxWithWrongContract  = errors.New("tx with wrong contract")
	ErrTxWithWrongInput     = errors.New("tx with wrong input data")
	ErrTxWithWrongLogData   = errors.New("tx with wrong log data")
	ErrLogIndexOutOfRange   = errors.New("log index out of range")
	ErrTxWithWrongTopics    = errors.New("tx with wrong log topics")
	ErrTxWithRemovedLog     = errors.New("tx with removed log")
	ErrTxIsAggregateTx      = errors.New("tx is aggregate tx")
	ErrWrongP2shBindAddress = errors.New("wrong p2sh bind address")
	ErrTxFuncHashMismatch   = errors.New("tx func hash mismatch")
	ErrDepositLogNotFound   = errors.New("deposit log not found or removed")
	ErrSwapoutLogNotFound   = errors.New("swapout log not found or removed")
	ErrUnknownPairID        = errors.New("unknown pair ID")
	ErrBindAddressMismatch  = errors.New("bind address mismatch")
	ErrTxWithoutReceipt     = errors.New("tx without receipt")
	ErrTxWithWrongReceipt   = errors.New("tx with wrong receipt")

	// errors should register
	ErrTxWithWrongMemo       = errors.New("tx with wrong memo")
	ErrTxWithWrongValue      = errors.New("tx with wrong value")
	ErrTxWithWrongSender     = errors.New("tx with wrong sender")
	ErrTxSenderNotRegistered = errors.New("tx sender not registered")
	ErrTxIncompatible        = errors.New("tx incompatible")
	ErrBindAddrIsContract    = errors.New("bind address is contract")
	ErrRPCQueryError         = errors.New("rpc query error")
	ErrTxWithWrongPath       = errors.New("swap trade tx with wrong path")
	ErrMissTokenConfig       = errors.New("miss token config")
)

// ShouldRegisterSwapForError return true if this error should record in database
func ShouldRegisterSwapForError(err error) bool {
	switch err {
	case nil,
		ErrTxWithWrongMemo,
		ErrTxWithWrongValue,
		ErrTxWithWrongSender,
		ErrTxSenderNotRegistered,
		ErrTxIncompatible,
		ErrBindAddrIsContract,
		ErrRPCQueryError:
		return true
	}
	return false
}

// ShouldRegisterRouterSwapForError return true if this error should record in database
func ShouldRegisterRouterSwapForError(err error) bool {
	switch err {
	case nil,
		ErrTxWithWrongValue,
		ErrTxWithWrongPath,
		ErrMissTokenConfig:
		return true
	}
	return false
}
