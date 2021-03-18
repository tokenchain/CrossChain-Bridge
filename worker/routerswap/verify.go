package routerswap

import (
	"github.com/anyswap/CrossChain-Bridge/mongodb"
	"github.com/anyswap/CrossChain-Bridge/params"
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
				logWorkerError("verify", "process router swap verify error", err, "chainid", swap.FromChainID, "txid", swap.TxID, "logIndex", swap.LogIndex)
			}
		}
		restInJob(restIntervalInVerifyJob)
	}
}

func processRouterSwapVerify(swap *mongodb.MgoSwap) (err error) {
	fromChainID := swap.FromChainID
	txid := swap.TxID
	logIndex := swap.LogIndex

	if params.IsSwapInBlacklist(fromChainID, swap.ToChainID, swap.TokenID) {
		err = tokens.ErrSwapInBlacklist
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.SwapInBlacklist, now(), err.Error())
	}

	bridge := router.GetBridgeByChainID(fromChainID)
	swapInfo, err := bridge.VerifyRouterSwapTx(txid, logIndex, false)
	if swapInfo == nil {
		return err
	}

	if swapInfo.Height != 0 && swapInfo.Height < bridge.ChainConfig.InitialHeight {
		err = tokens.ErrTxBeforeInitialHeight
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	}

	return updateSwapStatus(bridge, fromChainID, txid, logIndex, swapInfo, err)
}

func updateSwapStatus(bridge *router.Bridge, fromChainID, txid string, logIndex int, swapInfo *tokens.TxSwapInfo, err error) error {
	switch err {
	case tokens.ErrTxNotStable, tokens.ErrTxNotFound:
		return err
	case nil:
		if swapInfo.Value.Cmp(bridge.GetBigValueThreshold(swapInfo.Token)) > 0 {
			return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithBigValue, now(), "")
		}
		err = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxNotSwapped, now(), "")
		if err != nil {
			return err
		}
		return addInitialSwapResult(swapInfo, mongodb.MatchTxEmpty)
	case tokens.ErrTxWithWrongValue:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongValue, now(), err.Error())
	case tokens.ErrTxWithWrongPath:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongPath, now(), err.Error())
	case tokens.ErrMissTokenConfig:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.MissTokenConfig, now(), err.Error())
	default:
		return mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	}
}
