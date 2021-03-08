package rpcapi

import (
	"net/http"

	"github.com/anyswap/CrossChain-Bridge/internal/swapapi"
)

// VaultSwapAPI rpc api handler
type VaultSwapAPI struct{}

// VaultRegisterSwapArgs args
type VaultRegisterSwapArgs struct {
	ChainID string `json:"chainid"`
	TxID    string `json:"txid"`
}

// RegisterVaultSwap api
func (s *VaultSwapAPI) RegisterVaultSwap(r *http.Request, args *VaultRegisterSwapArgs, result *swapapi.MapIntResult) error {
	res, err := swapapi.RegisterVaultSwap(args.ChainID, args.TxID)
	if err == nil && res != nil {
		*result = *res
	}
	return err
}

// VaultGetSwapArgs args
type VaultGetSwapArgs struct {
	ChainID  string `json:"chainid"`
	TxID     string `json:"txid"`
	LogIndex string `json:"logindex"`
}

// GetVaultSwap api
func (s *VaultSwapAPI) GetVaultSwap(r *http.Request, args *VaultGetSwapArgs, result *swapapi.SwapInfo) error {
	res, err := swapapi.GetVaultSwap(args.ChainID, args.TxID, args.LogIndex)
	if err == nil && res != nil {
		*result = *res
	}
	return err
}

// VaultGetSwapHistoryArgs args
type VaultGetSwapHistoryArgs struct {
	ChainID string `json:"chainid"`
	Address string `json:"address"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// GetVaultSwapHistory api
func (s *VaultSwapAPI) GetVaultSwapHistory(r *http.Request, args *VaultGetSwapHistoryArgs, result *[]*swapapi.SwapInfo) error {
	res, err := swapapi.GetVaultSwapHistory(args.ChainID, args.Address, args.Offset, args.Limit)
	if err == nil && res != nil {
		*result = res
	}
	return err
}
