package rpcapi

import (
	"net/http"

	"github.com/anyswap/CrossChain-Bridge/internal/swapapi"
)

// RouterSwapAPI rpc api handler
type RouterSwapAPI struct{}

// RouterRegisterSwapArgs args
type RouterRegisterSwapArgs struct {
	ChainID string `json:"chainid"`
	TxID    string `json:"txid"`
}

// RegisterRouterSwap api
func (s *RouterSwapAPI) RegisterRouterSwap(r *http.Request, args *RouterRegisterSwapArgs, result *swapapi.MapIntResult) error {
	res, err := swapapi.RegisterRouterSwap(args.ChainID, args.TxID)
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
