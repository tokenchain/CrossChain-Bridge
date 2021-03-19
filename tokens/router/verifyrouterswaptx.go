package router

import (
	"bytes"
	"errors"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/types"
)

// router contract's log topics
var (
	// LogAnySwapOut(address token, address from, address to, uint amount, uint fromChainID, uint toChainID);
	LogAnySwapOutTopic = common.FromHex("0x97116cf6cd4f6412bb47914d6db18da9e16ab2142f543b86e207c24fbd16b23a")
	// LogAnySwapTradeTokensForTokens(address[] path, address from, address to, uint amountIn, uint amountOutMin, uint fromChainID, uint toChainID);
	LogAnySwapTradeTokensForTokensTopic = common.FromHex("0xfea6abdf4fd32f20966dff7619354cd82cd43dc78a3bee479f04c74dbfc585b3")
	// LogAnySwapTradeTokensForNative(address[] path, address from, address to, uint amountIn, uint amountOutMin, uint fromChainID, uint toChainID);
	LogAnySwapTradeTokensForNativeTopic = common.FromHex("0x278277e0209c347189add7bd92411973b5f6b8644f7ac62ea1be984ce993f8f4")

	errParseDataError = errors.New("parse data error")
)

// RegisterRouterSwapTx impl
func (b *Bridge) RegisterRouterSwapTx(txHash string, logIndex int) ([]*tokens.TxSwapInfo, []error) {
	commonInfo := &tokens.TxSwapInfo{}
	commonInfo.Hash = txHash // Hash

	txStatus := b.GetTransactionStatus(txHash)
	if txStatus.BlockHeight == 0 {
		return []*tokens.TxSwapInfo{commonInfo}, []error{tokens.ErrTxNotFound}
	}

	commonInfo.Height = txStatus.BlockHeight  // Height
	commonInfo.Timestamp = txStatus.BlockTime // Timestamp

	receipt, _ := txStatus.Receipt.(*types.RPCTxReceipt)
	err := b.verifyRouterSwapTxReceipt(commonInfo, receipt)
	if err != nil {
		return []*tokens.TxSwapInfo{commonInfo}, []error{err}
	}

	swapInfos := make([]*tokens.TxSwapInfo, 0)
	errs := make([]error, 0)
	startIndex, endIndex := 0, len(receipt.Logs)

	if logIndex != 0 {
		if logIndex >= endIndex || logIndex < 0 {
			return []*tokens.TxSwapInfo{commonInfo}, []error{tokens.ErrLogIndexOutOfRange}
		}
		startIndex = logIndex
		endIndex = logIndex + 1
	}

	for i := startIndex; i < endIndex; i++ {
		swapInfo := &tokens.TxSwapInfo{}
		*swapInfo = *commonInfo
		swapInfo.LogIndex = i // LogIndex
		err := b.verifyRouterSwapTxLog(swapInfo, receipt.Logs[i])
		if err == nil {
			err = b.checkRouterSwapInfo(swapInfo)
		}
		if tokens.ShouldRegisterRouterSwapForError(err) {
			swapInfos = append(swapInfos, swapInfo)
			errs = append(errs, err)
		}
	}

	return swapInfos, errs
}

// VerifyRouterSwapTx impl
func (b *Bridge) VerifyRouterSwapTx(txHash string, logIndex int, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	swapInfo := &tokens.TxSwapInfo{}
	swapInfo.Hash = txHash       // Hash
	swapInfo.LogIndex = logIndex // LogIndex

	txStatus := b.GetTransactionStatus(txHash)
	if txStatus.BlockHeight == 0 {
		return swapInfo, tokens.ErrTxNotFound
	}

	swapInfo.Height = txStatus.BlockHeight  // Height
	swapInfo.Timestamp = txStatus.BlockTime // Timestamp

	if !allowUnstable && txStatus.Confirmations < b.ChainConfig.Confirmations {
		return swapInfo, tokens.ErrTxNotStable
	}

	receipt, _ := txStatus.Receipt.(*types.RPCTxReceipt)
	err := b.verifyRouterSwapTxReceipt(swapInfo, receipt)
	if err != nil {
		return swapInfo, err
	}

	if logIndex >= len(receipt.Logs) {
		return swapInfo, tokens.ErrLogIndexOutOfRange
	}

	err = b.verifyRouterSwapTxLog(swapInfo, receipt.Logs[logIndex])
	if err != nil {
		return swapInfo, err
	}

	err = b.checkRouterSwapInfo(swapInfo)
	if err != nil {
		return swapInfo, err
	}

	if !allowUnstable {
		log.Debug("verify router swap tx stable pass",
			"from", swapInfo.From, "to", swapInfo.To, "bind", swapInfo.Bind, "value", swapInfo.Value,
			"txid", txHash, "logIndex", logIndex, "height", swapInfo.Height, "timestamp", swapInfo.Timestamp,
			"fromChainID", swapInfo.FromChainID, "toChainID", swapInfo.ToChainID,
			"token", swapInfo.Token, "tokenID", swapInfo.TokenID,
			"forNative", swapInfo.ForNative, "forUnderlying", swapInfo.ForUnderlying)
	}

	return swapInfo, nil
}

func (b *Bridge) checkRouterSwapInfo(swapInfo *tokens.TxSwapInfo) error {
	tokenCfg := b.GetTokenConfig(swapInfo.Token)
	if tokenCfg == nil {
		return tokens.ErrMissTokenConfig
	}
	if !CheckTokenSwapValue(tokenCfg, swapInfo.Value) {
		return tokens.ErrTxWithWrongValue
	}
	dstBridge := GetBridgeByChainID(swapInfo.ToChainID.String())
	if dstBridge == nil {
		return tokens.ErrNoBridgeForChainID
	}
	if !dstBridge.IsValidAddress(swapInfo.Bind) {
		log.Debug("wrong bind address in router swap", "txid", swapInfo.Hash, "logIndex", swapInfo.LogIndex, "bind", swapInfo.Bind)
		return tokens.ErrTxWithWrongMemo
	}
	return nil
}

func (b *Bridge) verifyRouterSwapTxReceipt(swapInfo *tokens.TxSwapInfo, receipt *types.RPCTxReceipt) (err error) {
	if receipt == nil || *receipt.Status != 1 {
		return tokens.ErrTxWithWrongReceipt
	}

	if receipt.Recipient == nil {
		return tokens.ErrTxWithWrongContract
	}

	routerContract := b.ChainConfig.RouterContract
	txRecipient := strings.ToLower(receipt.Recipient.String())
	if !common.IsEqualIgnoreCase(txRecipient, routerContract) {
		return tokens.ErrTxWithWrongContract
	}

	swapInfo.TxTo = txRecipient                            // TxTo
	swapInfo.To = txRecipient                              // To
	swapInfo.From = strings.ToLower(receipt.From.String()) // From
	return nil
}

func (b *Bridge) verifyRouterSwapTxLog(swapInfo *tokens.TxSwapInfo, rlog *types.RPCLog) (err error) {
	if rlog.Removed != nil && *rlog.Removed {
		return tokens.ErrTxWithRemovedLog
	}

	logTopic := rlog.Topics[0].Bytes()
	switch {
	case bytes.Equal(logTopic, LogAnySwapOutTopic):
		err = b.parseRouterSwapoutTxLog(swapInfo, rlog)
	case bytes.Equal(logTopic, LogAnySwapTradeTokensForTokensTopic):
		err = b.parseRouterSwapTradeTxLog(swapInfo, rlog, false)
	case bytes.Equal(logTopic, LogAnySwapTradeTokensForNativeTopic):
		err = b.parseRouterSwapTradeTxLog(swapInfo, rlog, true)
	default:
		return tokens.ErrSwapoutLogNotFound
	}
	if err != nil {
		log.Debug(b.ChainConfig.BlockChain+" b.verifyRouterSwapTxLog fail", "tx", swapInfo.Hash, "logIndex", rlog.Index, "err", err)
		return err
	}
	tokenCfg := b.GetTokenConfig(swapInfo.Token)
	if tokenCfg == nil {
		return tokens.ErrMissTokenConfig
	}
	swapInfo.TokenID = tokenCfg.ID
	return nil
}

func (b *Bridge) parseRouterSwapoutTxLog(swapInfo *tokens.TxSwapInfo, rlog *types.RPCLog) error {
	logTopics := rlog.Topics
	if len(logTopics) != 4 {
		return tokens.ErrTxWithWrongTopics
	}
	logData := *rlog.Data
	if len(logData) != 128 {
		return errParseDataError
	}
	swapInfo.Token = common.BytesToAddress(logTopics[1].Bytes()).String()
	swapInfo.From = common.BytesToAddress(logTopics[2].Bytes()).String()
	swapInfo.Bind = common.BytesToAddress(logTopics[3].Bytes()).String()
	swapInfo.Value = common.GetBigInt(logData, 0, 32)
	swapInfo.FromChainID = common.GetBigInt(logData, 32, 32)
	swapInfo.ToChainID = common.GetBigInt(logData, 64, 32)
	swapInfo.ForUnderlying = common.GetBigInt(logData, 96, 32).Sign() != 0
	return nil
}

func (b *Bridge) parseRouterSwapTradeTxLog(swapInfo *tokens.TxSwapInfo, rlog *types.RPCLog, forNative bool) error {
	logTopics := rlog.Topics
	if len(logTopics) != 3 {
		return tokens.ErrTxWithWrongTopics
	}
	logData := *rlog.Data
	if len(logData) < 192 {
		return errParseDataError
	}
	swapInfo.ForNative = forNative
	swapInfo.From = common.BytesToAddress(logTopics[1].Bytes()).String()
	swapInfo.Bind = common.BytesToAddress(logTopics[2].Bytes()).String()
	path, err := ParseAddressSliceInData(logData, 0)
	if err != nil {
		return err
	}
	if len(swapInfo.Path) < 2 {
		return tokens.ErrTxWithWrongPath
	}
	swapInfo.Value = common.GetBigInt(logData, 32, 32)
	swapInfo.AmountOutMin = common.GetBigInt(logData, 64, 32)
	swapInfo.FromChainID = common.GetBigInt(logData, 96, 32)
	swapInfo.ToChainID = common.GetBigInt(logData, 128, 32)

	swapInfo.Token = path[0]
	swapInfo.Path = path[1:]
	return b.chekcAndAmendSwapTradePath(swapInfo)
}

// amend trade path [0] if missing,
// then check path exists in pairs of dest chain
func (b *Bridge) chekcAndAmendSwapTradePath(swapInfo *tokens.TxSwapInfo) error {
	dstBridge := GetBridgeByChainID(swapInfo.ToChainID.String())
	if dstBridge == nil {
		return tokens.ErrNoBridgeForChainID
	}
	tokenID := swapInfo.TokenID
	dstChainID := dstBridge.ChainConfig.ChainID
	peerToken := GetPeerToken(tokenID, dstChainID)
	if peerToken == "" {
		return tokens.ErrMissTokenConfig
	}
	tokenCfg := dstBridge.GetTokenConfig(peerToken)
	if tokenCfg == nil {
		return tokens.ErrMissTokenConfig
	}
	path := swapInfo.Path
	if common.HexToAddress(path[0]) != common.HexToAddress(peerToken) {
		path = append([]string{peerToken}, path...)
		swapInfo.Path = path
	}
	if len(path) < 2 {
		return tokens.ErrTxWithWrongPath
	}
	factory := b.ChainConfig.GetRouterFactory()
	if factory == "" {
		return tokens.ErrTxWithWrongPath
	}
	for i := 1; i < len(path); i++ {
		pairs, err := b.GetPairFor(factory, path[i-1], path[i])
		if err != nil || pairs == "" {
			return tokens.ErrTxWithWrongPath
		}
	}
	return nil
}
