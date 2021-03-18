package rpcapi

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/anyswap/CrossChain-Bridge/admin"
	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/internal/swapapi"
	"github.com/anyswap/CrossChain-Bridge/mongodb"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/worker/routerswap"
)

// RouterSwapAPI rpc api handler
type RouterSwapAPI struct{}

// RouterRegisterSwapArgs args
type RouterRegisterSwapArgs struct {
	ChainID  string `json:"chainid"`
	TxID     string `json:"txid"`
	LogIndex string `json:"logindex"`
}

// RegisterRouterSwap api
func (s *RouterSwapAPI) RegisterRouterSwap(r *http.Request, args *RouterRegisterSwapArgs, result *swapapi.MapIntResult) error {
	res, err := swapapi.RegisterRouterSwap(args.ChainID, args.TxID, args.LogIndex)
	if err == nil && res != nil {
		*result = *res
	}
	return err
}

// RouterGetSwapArgs args
type RouterGetSwapArgs struct {
	ChainID  string `json:"chainid"`
	TxID     string `json:"txid"`
	LogIndex string `json:"logindex"`
}

// GetRouterSwap api
func (s *RouterSwapAPI) GetRouterSwap(r *http.Request, args *RouterGetSwapArgs, result *swapapi.SwapInfo) error {
	res, err := swapapi.GetRouterSwap(args.ChainID, args.TxID, args.LogIndex)
	if err == nil && res != nil {
		*result = *res
	}
	return err
}

// RouterGetSwapHistoryArgs args
type RouterGetSwapHistoryArgs struct {
	ChainID string `json:"chainid"`
	Address string `json:"address"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// GetRouterSwapHistory api
func (s *RouterSwapAPI) GetRouterSwapHistory(r *http.Request, args *RouterGetSwapHistoryArgs, result *[]*swapapi.SwapInfo) error {
	res, err := swapapi.GetRouterSwapHistory(args.ChainID, args.Address, args.Offset, args.Limit)
	if err == nil && res != nil {
		*result = res
	}
	return err
}

// AdminCall admin call
func (s *RouterSwapAPI) AdminCall(r *http.Request, rawTx, result *string) (err error) {
	if !params.HasRouterAdmin() {
		return fmt.Errorf("no admin is configed")
	}
	tx, err := admin.DecodeTransaction(*rawTx)
	if err != nil {
		return err
	}
	sender, args, err := admin.VerifyTransaction(tx)
	if err != nil {
		return err
	}
	if !params.IsRouterAdmin(sender.String()) {
		return fmt.Errorf("sender %v is not admin", sender.String())
	}
	return doRouterAdminCall(args, result)
}

func doRouterAdminCall(args *admin.CallArgs, result *string) error {
	switch args.Method {
	case passbigvalueCmd:
		return routerPassBigValue(args, result)
	case reswapCmd:
		return routerReswap(args, result)
	case replaceswapCmd:
		return routerReplaceSwap(args, result)
	default:
		return fmt.Errorf("unknown admin method '%v'", args.Method)
	}
}

func getKeys(args *admin.CallArgs, startPos int) (chainID, txid string, logIndex int, err error) {
	if len(args.Params) < startPos+3 {
		err = fmt.Errorf("wrong number of params, have %v want at least %v", len(args.Params), startPos+3)
		return
	}
	chainID = args.Params[startPos]
	if _, err = common.GetBigIntFromStr(chainID); err != nil || chainID == "" {
		err = fmt.Errorf("wrong chain id '%v'", chainID)
		return
	}
	txid = args.Params[startPos+1]
	if !common.IsHexHash(txid) {
		err = fmt.Errorf("wrong tx id '%v'", txid)
		return
	}
	logIndexStr := args.Params[startPos+2]
	logIndex, err = common.GetIntFromStr(logIndexStr)
	if err != nil {
		err = fmt.Errorf("wrong log index '%v'", logIndexStr)
	}
	return
}

func getGasPrice(args *admin.CallArgs, startPos int) (gasPrice *big.Int, err error) {
	if len(args.Params) < startPos+1 {
		err = fmt.Errorf("wrong number of params, have %v want at least %v", len(args.Params), startPos+3)
		return
	}
	gasPriceStr := args.Params[startPos]
	if gasPrice, err = common.GetBigIntFromStr(gasPriceStr); err != nil {
		err = fmt.Errorf("wrong gas price '%v'", gasPriceStr)
	}
	return
}

func routerPassBigValue(args *admin.CallArgs, result *string) (err error) {
	chainID, txid, logIndex, err := getKeys(args, 0)
	if err != nil {
		return err
	}
	err = mongodb.RouterAdminPassBigValue(chainID, txid, logIndex)
	if err != nil {
		return err
	}
	*result = successReuslt
	return nil
}

func routerReswap(args *admin.CallArgs, result *string) (err error) {
	chainID, txid, logIndex, err := getKeys(args, 0)
	if err != nil {
		return err
	}
	err = mongodb.RouterAdminReswap(chainID, txid, logIndex)
	if err != nil {
		return err
	}
	*result = successReuslt
	return nil
}

func routerReplaceSwap(args *admin.CallArgs, result *string) (err error) {
	chainID, txid, logIndex, err := getKeys(args, 0)
	if err != nil {
		return err
	}
	gasPrice, err := getGasPrice(args, 3)
	if err != nil {
		return err
	}
	res, err := mongodb.FindRouterSwapResult(chainID, txid, logIndex)
	if err != nil {
		return err
	}
	err = routerswap.ReplaceRouterSwap(res, gasPrice)
	if err != nil {
		return err
	}
	*result = successReuslt
	return nil
}
