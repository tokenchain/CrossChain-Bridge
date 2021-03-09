package routerswap

import (
	"container/ring"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/mongodb"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

var (
	swapRing        *ring.Ring
	swapRingLock    sync.RWMutex
	swapRingMaxSize = 1000

	swapChanSize          = 10
	routerSwapTaskChanMap = make(map[string]chan *tokens.BuildTxArgs)

	errAlreadySwapped = errors.New("already swapped")
)

// StartSwapJob swap job
func StartSwapJob() {
	for _, pairCfg := range tokens.GetTokenPairsConfig() {
		AddSwapJob(pairCfg)
	}
}

// AddSwapJob add swap job
func AddSwapJob(pairCfg *tokens.TokenPairConfig) {
	chainID := strings.ToLower(pairCfg.PairID)
	if _, exist := routerSwapTaskChanMap[chainID]; !exist {
		routerSwapTaskChanMap[chainID] = make(chan *tokens.BuildTxArgs, swapChanSize)
		go processSwapTask(routerSwapTaskChanMap[chainID])
	}

	go startRouterSwapJob(chainID)
}

func startRouterSwapJob(chainID string) {
	logWorker("swap", "start router swap job")
	for {
		res, err := findRouterSwapToSwap(chainID)
		if err != nil {
			logWorkerError("swap", "find out router swap error", err)
		}
		if len(res) > 0 {
			logWorker("swap", "find out router swap", "count", len(res))
		}
		for _, swap := range res {
			err = processRouterSwap(swap)
			switch err {
			case nil, errAlreadySwapped:
			default:
				logWorkerError("swap", "process router swap error", err, "chainID", chainID, "txid", swap.TxID, "logIndex", swap.LogIndex)
			}
		}
		restInJob(restIntervalInDoSwapJob)
	}
}

func findRouterSwapToSwap(chainID string) ([]*mongodb.MgoSwap, error) {
	status := mongodb.TxNotSwapped
	septime := getSepTimeInFind(maxDoSwapLifetime)
	return mongodb.FindRouterSwapsWithChainIDAndStatus(chainID, status, septime)
}

func isSwapInBlacklist(swap *mongodb.MgoSwapResult) (isBlacked bool, err error) {
	isBlacked, err = mongodb.QueryBlacklist(swap.From, swap.PairID)
	if err != nil {
		return isBlacked, err
	}
	if !isBlacked && swap.Bind != swap.From {
		isBlacked, err = mongodb.QueryBlacklist(swap.Bind, swap.PairID)
		if err != nil {
			return isBlacked, err
		}
	}
	return isBlacked, nil
}

func processRouterSwap(swap *mongodb.MgoSwap) (err error) {
	chainID := swap.FromChainID
	txid := swap.TxID
	logIndex := swap.LogIndex
	bind := swap.Bind

	res, err := mongodb.FindRouterSwapResult(chainID, txid, logIndex)
	if err != nil {
		return err
	}

	logWorker("swap", "start process router swap", "chainID", chainID, "txid", txid, "logIndex", logIndex, "status", swap.Status, "value", res.Value)

	fromTokenCfg, toTokenCfg := tokens.GetTokenConfigsByDirection(chainID, true) // TODO
	if fromTokenCfg == nil || toTokenCfg == nil {
		logWorkerTrace("swap", "swap is not configed", "chainID", chainID, "txid", txid)
		return nil
	}
	if fromTokenCfg.DisableSwap {
		logWorkerTrace("swap", "swap is disabled", "chainID", chainID, "txid", txid)
		return nil
	}
	isBlacked, err := isSwapInBlacklist(res)
	if err != nil {
		return err
	}
	if isBlacked {
		logWorkerTrace("swap", "address is in blacklist", "chainID", chainID, "txid", txid, "logIndex", logIndex)
		err = tokens.ErrAddressIsInBlacklist
		_ = mongodb.UpdateRouterSwapStatus(chainID, txid, logIndex, mongodb.SwapInBlacklist, now(), err.Error())
		return nil
	}

	err = preventReswap(res)
	if err != nil {
		return err
	}

	value, err := common.GetBigIntFromStr(res.Value)
	if err != nil {
		return fmt.Errorf("wrong value %v", res.Value)
	}

	args := &tokens.BuildTxArgs{
		SwapInfo: tokens.SwapInfo{
			PairID:   chainID,
			SwapID:   txid,
			SwapType: tokens.RouterSwapType,
			TxType:   tokens.SwapTxType(swap.TxType),
			Bind:     bind,
		},
		From:        toTokenCfg.DcrmAddress,
		OriginValue: value,
	}

	return dispatchSwapTask(args)
}

func preventReswap(res *mongodb.MgoSwapResult) (err error) {
	err = processNonEmptySwapResult(res)
	if err != nil {
		return err
	}
	return processHistory(res)
}

func processNonEmptySwapResult(res *mongodb.MgoSwapResult) error {
	if res.SwapTx == "" {
		return nil
	}
	chainID := res.FromChainID
	txid := res.TxID
	logIndex := res.LogIndex
	_ = mongodb.UpdateRouterSwapStatus(chainID, txid, logIndex, mongodb.TxProcessed, now(), "")
	if res.Status != mongodb.MatchTxEmpty {
		return errAlreadySwapped
	}
	resBridge := tokens.GetCrossChainBridgeByChainID(res.ToChainID)
	if _, err := resBridge.GetTransaction(res.SwapTx); err == nil {
		return errAlreadySwapped
	}
	return nil
}

func processHistory(res *mongodb.MgoSwapResult) error {
	chainID := res.FromChainID
	txid := res.TxID
	logIndex := res.LogIndex
	history := getSwapHistory(chainID, txid, logIndex)
	if history == nil {
		return nil
	}
	if res.Status == mongodb.MatchTxFailed {
		history.txid = "" // mark ineffective
		return nil
	}
	resBridge := tokens.GetCrossChainBridgeByChainID(res.ToChainID)
	if _, err := resBridge.GetTransaction(history.matchTx); err == nil {
		matchTx := &MatchTx{
			SwapTx:    history.matchTx,
			SwapValue: tokens.CalcSwappedValue(chainID, history.value, true).String(), // TODO
			SwapNonce: history.nonce,
		}
		_ = updateRouterSwapResult(chainID, txid, logIndex, matchTx)
		logWorker("swap", "ignore swapped router swap", "chainID", chainID, "txid", txid, "matchTx", history.matchTx)
		return errAlreadySwapped
	}
	return nil
}

func dispatchSwapTask(args *tokens.BuildTxArgs) error {
	switch args.SwapType {
	case tokens.RouterSwapType:
		swapChan, exist := routerSwapTaskChanMap[args.FromChainID.String()]
		if !exist {
			return fmt.Errorf("no swapout task channel for chainID '%v'", args.FromChainID)
		}
		swapChan <- args
	default:
		return fmt.Errorf("wrong swap type '%v'", args.SwapType.String())
	}
	logWorker("doSwap", "dispatch router swap task", "chainID", args.FromChainID, "txid", args.SwapID, "logIndex", args.LogIndex, "value", args.OriginValue)
	return nil
}

func processSwapTask(swapChan <-chan *tokens.BuildTxArgs) {
	for {
		args := <-swapChan
		err := doSwap(args)
		switch err {
		case nil, errAlreadySwapped:
		default:
			logWorkerError("doSwap", "process router swap failed", err, "chainID", args.FromChainID, "txid", args.SwapID, "value", args.OriginValue)
		}
	}
}

func doSwap(args *tokens.BuildTxArgs) (err error) {
	chainID := args.FromChainID.String()
	txid := args.SwapID
	logIndex := args.LogIndex
	originValue := args.OriginValue

	resBridge := tokens.GetCrossChainBridgeByChainID(args.ToChainID.String())

	res, err := mongodb.FindRouterSwapResult(chainID, txid, logIndex)
	if err != nil {
		return err
	}
	err = preventReswap(res)
	if err != nil {
		return err
	}

	logWorker("doSwap", "start to process", "chainID", chainID, "txid", txid, "logIndex", logIndex, "value", originValue)

	rawTx, err := resBridge.BuildRawTransaction(args)
	if err != nil {
		logWorkerError("doSwap", "build tx failed", err, "chainID", chainID, "txid", txid, "logIndex", logIndex)
		return err
	}

	var signedTx interface{}
	var txHash string
	tokenCfg := resBridge.GetTokenConfig(chainID)
	if tokenCfg.GetDcrmAddressPrivateKey() != nil {
		signedTx, txHash, err = resBridge.SignTransaction(rawTx, chainID)
	} else {
		signedTx, txHash, err = dcrmSignTransaction(resBridge, rawTx, args.GetExtraArgs())
	}
	if err != nil {
		logWorkerError("doSwap", "sign tx failed", err, "chainID", chainID, "txid", txid, "logIndex", logIndex)
		return err
	}

	swapTxNonce := args.GetTxNonce()

	var existsInOld bool
	var oldSwapTxs []string
	for _, oldSwapTx := range res.OldSwapTxs {
		if oldSwapTx == txHash {
			existsInOld = true
			break
		}
	}
	if !existsInOld {
		oldSwapTxs = res.OldSwapTxs
		oldSwapTxs = append(oldSwapTxs, txHash)
	}

	// update database before sending transaction
	addSwapHistory(chainID, txid, logIndex, originValue, txHash, swapTxNonce)
	matchTx := &MatchTx{
		SwapTx:     txHash,
		OldSwapTxs: oldSwapTxs,
		SwapValue:  tokens.CalcSwappedValue(chainID, originValue, true).String(), // TODO
		SwapNonce:  swapTxNonce,
	}
	err = updateRouterSwapResult(chainID, txid, logIndex, matchTx)
	if err != nil {
		logWorkerError("doSwap", "update router swap result failed", err, "chainID", chainID, "txid", txid, "logIndex", logIndex)
		return err
	}

	err = mongodb.UpdateRouterSwapStatus(chainID, txid, logIndex, mongodb.TxProcessed, now(), "")
	if err != nil {
		logWorkerError("doSwap", "update router swap status failed", err, "chainID", chainID, "txid", txid, "logIndex", logIndex)
		return err
	}

	return sendSignedTransaction(resBridge, signedTx, chainID, txid, logIndex, false)
}

type swapInfo struct {
	chainID  string
	txid     string
	logIndex int
	value    *big.Int
	matchTx  string
	nonce    uint64
}

func addSwapHistory(chainID, txid string, logIndex int, value *big.Int, matchTx string, nonce uint64) {
	// Create the new item as its own ring
	item := ring.New(1)
	item.Value = &swapInfo{
		chainID:  chainID,
		txid:     txid,
		logIndex: logIndex,
		value:    value,
		matchTx:  matchTx,
		nonce:    nonce,
	}

	swapRingLock.Lock()
	defer swapRingLock.Unlock()

	if swapRing == nil {
		swapRing = item
	} else {
		if swapRing.Len() == swapRingMaxSize {
			swapRing = swapRing.Move(-1)
			swapRing.Unlink(1)
			swapRing = swapRing.Move(1)
		}
		swapRing.Move(-1).Link(item)
	}
}

func getSwapHistory(chainID, txid string, logIndex int) *swapInfo {
	swapRingLock.RLock()
	defer swapRingLock.RUnlock()

	if swapRing == nil {
		return nil
	}

	r := swapRing
	for i := 0; i < r.Len(); i++ {
		item := r.Value.(*swapInfo)
		if item.txid == txid && item.chainID == chainID && item.logIndex == logIndex {
			return item
		}
		r = r.Prev()
	}

	return nil
}
