package routerswap

import (
	"github.com/anyswap/CrossChain-Bridge/mongodb"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/router"
)

// StartVerifyJob verify job
func StartVerifyJob() {
	logWorker("verify", "start router swap verify job")
	for {
		septime := getSepTimeInFind(maxVerifyLifetime)
		res, err := mongodb.FindRouterSwapsWithStatus(mongodb.TxNotStable, septime)
		if err != nil {
			logWorkerError("verify", "find router swap error", err)
		}
		if len(res) > 0 {
			logWorker("verify", "find router swap to verify", "count", len(res))
		}
		for _, swap := range res {
			err = processRouterSwapVerify(swap)
			switch err {
			case nil, tokens.ErrTxNotStable, tokens.ErrTxNotFound:
			default:
				logWorkerError("verify", "process router swap verify error", err, "txid", swap.TxID, "logIndex", swap.LogIndex)
			}
		}
		restInJob(restIntervalInVerifyJob)
	}
}

func isInBlacklist(swapInfo *tokens.TxSwapInfo) (isBlacked bool, err error) {
	isBlacked, err = mongodb.QueryBlacklist(swapInfo.From, swapInfo.PairID)
	if err != nil {
		return isBlacked, err
	}
	if !isBlacked && swapInfo.Bind != swapInfo.From {
		isBlacked, err = mongodb.QueryBlacklist(swapInfo.Bind, swapInfo.PairID)
		if err != nil {
			return isBlacked, err
		}
	}
	return isBlacked, nil
}

func processRouterSwapVerify(swap *mongodb.MgoSwap) (err error) {
	fromChainID := swap.FromChainID
	txid := swap.TxID
	logIndex := swap.LogIndex

	bridge := router.GetBridgeByChainID(fromChainID)
	swapInfo, err := bridge.VerifyRouterSwapTx(txid, logIndex, false)
	if swapInfo == nil {
		return err
	}

	if swapInfo.Height != 0 && swapInfo.Height < *bridge.ChainConfig.InitialHeight {
		err = tokens.ErrTxBeforeInitialHeight
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	}
	isBlacked, errf := isInBlacklist(swapInfo)
	if errf != nil {
		return errf
	}
	if isBlacked {
		err = tokens.ErrAddressIsInBlacklist
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.SwapInBlacklist, now(), err.Error())
	}
	return updateSwapStatus(bridge, fromChainID, txid, logIndex, swapInfo, err)
}

func updateSwapStatus(bridge *router.Bridge, fromChainID, txid string, logIndex int, swapInfo *tokens.TxSwapInfo, err error) error {
	resultStatus := mongodb.MatchTxEmpty

	switch err {
	case tokens.ErrTxNotStable, tokens.ErrTxNotFound:
		return err
	case nil:
		status := mongodb.TxNotSwapped
		if swapInfo.Value.Cmp(bridge.GetBigValueThreshold(swapInfo.Token)) > 0 {
			status = mongodb.TxWithBigValue
			resultStatus = mongodb.TxWithBigValue
		}
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, status, now(), "")
	case tokens.ErrTxWithWrongPath:
		resultStatus = mongodb.TxWithWrongPath
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongPath, now(), err.Error())
	case tokens.ErrTxWithWrongMemo:
		resultStatus = mongodb.TxWithWrongMemo
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongMemo, now(), err.Error())
	case tokens.ErrBindAddrIsContract:
		resultStatus = mongodb.BindAddrIsContract
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.BindAddrIsContract, now(), err.Error())
	case tokens.ErrTxWithWrongValue:
		resultStatus = mongodb.TxWithWrongValue
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongValue, now(), err.Error())
	case tokens.ErrTxSenderNotRegistered:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxSenderNotRegistered, now(), err.Error())
	case tokens.ErrTxWithWrongSender:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongSender, now(), err.Error())
	case tokens.ErrTxIncompatible:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxIncompatible, now(), err.Error())
	case tokens.ErrTxWithWrongReceipt, tokens.ErrBindAddressMismatch:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	case tokens.ErrRPCQueryError:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.RPCQueryError, now(), err.Error())
	default:
		logWorkerWarn("verify", "maybe not considered tx verify error", "err", err)
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	}

	if err != nil {
		logWorkerError("verify", "update router swap status", err, "chainid", fromChainID, "txid", txid, "logIndex", swapInfo.LogIndex)
		return err
	}
	return addInitialSwapResult(swapInfo, resultStatus)
}
