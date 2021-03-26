package main

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/cmd/utils"
	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/tokens/router"
	"github.com/urfave/cli/v2"
)

var (
	configCommand = &cli.Command{
		Name:  "config",
		Usage: "config router swap",
		Flags: utils.CommonLogFlags,
		Description: `
config router swap
`,
		Subcommands: []*cli.Command{
			{
				Name:   "genSetChainConfigData",
				Usage:  "generate setChainConfig input data",
				Action: genSetChainConfigData,
				Flags: []cli.Flag{
					cChainIDFlag,
					cBlockChainFlag,
					cRouterContractFlag,
					cConfirmationsFlag,
					cInitialHeightFlag,
					cWaitTimeToReplaceFlag,
					cMaxReplaceCountFlag,
					cSwapDeadlineOffsetFlag,
					cPlusGasPricePercentageFlag,
					cMaxGasPriceFluctPercentFlag,
					cDefaultGasLimitFlag,
				},
				Description: `
generate ChainConfig json marshal data
`,
			},
			{
				Name:   "genSetTokenConfigData",
				Usage:  "generate setTokenConfig input data",
				Action: genSetTokenConfigData,
				Flags: []cli.Flag{
					cChainIDFlag,
					cTokenIDFlag,
					cDecimalsFlag,
					cContractAddressFlag,
					cContractVersionFlag,
					cMaximumSwapFlag,
					cMinimumSwapFlag,
					cBigValueThresholdFlag,
					cSwapFeeRateFlag,
					cMaximumSwapFeeFlag,
					cMinimumSwapFeeFlag,
				},
				Description: `
generate TokenConfig json marshal data
`,
			},
		},
	}

	// --------- chain config -------------------

	cChainIDFlag = &cli.StringFlag{
		Name:  "c.ChainID",
		Usage: "chain config (require)",
	}

	cBlockChainFlag = &cli.StringFlag{
		Name:  "c.BlockChain",
		Usage: "chain config (require)",
	}

	cRouterContractFlag = &cli.StringFlag{
		Name:  "c.RouterContract",
		Usage: "chain config (require)",
	}

	cConfirmationsFlag = &cli.Uint64Flag{
		Name:  "c.Confirmations",
		Usage: "chain config (require)",
	}

	cInitialHeightFlag = &cli.Uint64Flag{
		Name:  "c.InitialHeight",
		Usage: "chain config",
	}

	cWaitTimeToReplaceFlag = &cli.Int64Flag{
		Name:  "c.WaitTimeToReplace",
		Usage: "chain config",
		Value: 900,
	}

	cMaxReplaceCountFlag = &cli.Int64Flag{
		Name:  "c.MaxReplaceCount",
		Usage: "chain config",
		Value: 20,
	}

	cSwapDeadlineOffsetFlag = &cli.Int64Flag{
		Name:  "c.SwapDeadlineOffset",
		Usage: "chain config",
		Value: 36000,
	}

	cPlusGasPricePercentageFlag = &cli.Uint64Flag{
		Name:  "c.PlusGasPricePercentage",
		Usage: "chain config",
	}

	cMaxGasPriceFluctPercentFlag = &cli.Uint64Flag{
		Name:  "c.MaxGasPriceFluctPercent",
		Usage: "chain config",
	}

	cDefaultGasLimitFlag = &cli.Uint64Flag{
		Name:  "c.DefaultGasLimit",
		Usage: "chain config",
		Value: 90000,
	}

	// --------- token config -------------------

	cTokenIDFlag = &cli.StringFlag{
		Name:  "c.TokenID",
		Usage: "token config (require)",
	}

	cDecimalsFlag = &cli.IntFlag{
		Name:  "c.Decimals",
		Usage: "token config",
		Value: 18,
	}

	cContractAddressFlag = &cli.StringFlag{
		Name:  "c.ContractAddress",
		Usage: "token config (require)",
	}

	cContractVersionFlag = &cli.Uint64Flag{
		Name:  "c.ContractVersion",
		Usage: "token config (require)",
	}

	cMaximumSwapFlag = &cli.Float64Flag{
		Name:  "c.MaximumSwap",
		Usage: "token config (require)",
	}

	cMinimumSwapFlag = &cli.Float64Flag{
		Name:  "c.MinimumSwap",
		Usage: "token config (require)",
	}

	cBigValueThresholdFlag = &cli.Float64Flag{
		Name:  "c.BigValueThreshold",
		Usage: "token config (require)",
	}

	cSwapFeeRateFlag = &cli.Float64Flag{
		Name:  "c.SwapFeeRate",
		Usage: "token config (require)",
	}

	cMaximumSwapFeeFlag = &cli.Float64Flag{
		Name:  "c.MaximumSwapFee",
		Usage: "token config",
	}

	cMinimumSwapFeeFlag = &cli.Float64Flag{
		Name:  "c.MinimumSwapFee",
		Usage: "token config",
	}
)

func genSetChainConfigData(ctx *cli.Context) error {
	chainCfg := &router.ChainConfig{
		ChainID:                 ctx.String(cChainIDFlag.Name),
		BlockChain:              ctx.String(cBlockChainFlag.Name),
		RouterContract:          ctx.String(cRouterContractFlag.Name),
		Confirmations:           ctx.Uint64(cConfirmationsFlag.Name),
		InitialHeight:           ctx.Uint64(cInitialHeightFlag.Name),
		WaitTimeToReplace:       ctx.Int64(cWaitTimeToReplaceFlag.Name),
		MaxReplaceCount:         ctx.Int(cMaxReplaceCountFlag.Name),
		SwapDeadlineOffset:      ctx.Int64(cSwapDeadlineOffsetFlag.Name),
		PlusGasPricePercentage:  ctx.Uint64(cPlusGasPricePercentageFlag.Name),
		MaxGasPriceFluctPercent: ctx.Uint64(cMaxGasPriceFluctPercentFlag.Name),
		DefaultGasLimit:         ctx.Uint64(cDefaultGasLimitFlag.Name),
	}
	err := chainCfg.CheckConfig()
	if err != nil {
		return err
	}
	jsdata, err := json.MarshalIndent(chainCfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("chain config struct is", string(jsdata))
	funcHash := common.FromHex("0xdefb3a0d")
	configData := router.PackData(
		chainCfg.BlockChain,
		common.HexToAddress(chainCfg.RouterContract),
		chainCfg.Confirmations,
		chainCfg.InitialHeight,
		chainCfg.WaitTimeToReplace,
		chainCfg.MaxReplaceCount,
		chainCfg.SwapDeadlineOffset,
		chainCfg.PlusGasPricePercentage,
		chainCfg.MaxGasPriceFluctPercent,
		chainCfg.DefaultGasLimit,
	)
	chainID, _ := new(big.Int).SetString(chainCfg.ChainID, 0)
	inputData := router.PackDataWithFuncHash(funcHash, chainID)
	inputData = append(inputData, common.LeftPadBytes([]byte{0x40}, 32)...)
	inputData = append(inputData, configData...)
	fmt.Println("set chain config input data is", common.ToHex(inputData))
	return nil
}

func genSetTokenConfigData(ctx *cli.Context) error {
	chainIDStr := ctx.String(cChainIDFlag.Name)
	chainID, err := common.GetBigIntFromStr(chainIDStr)
	if err != nil {
		return fmt.Errorf("wrong chainID '%v'", chainIDStr)
	}
	decimalsVal := ctx.Int(cDecimalsFlag.Name)
	if decimalsVal < 0 || decimalsVal > 256 {
		return fmt.Errorf("wrong decimals '%v'", decimalsVal)
	}
	decimals := uint8(decimalsVal)
	tokenCfg := &router.TokenConfig{
		TokenID:           ctx.String(cTokenIDFlag.Name),
		Decimals:          decimals,
		ContractAddress:   ctx.String(cContractAddressFlag.Name),
		ContractVersion:   ctx.Uint64(cContractVersionFlag.Name),
		MaximumSwap:       ctx.Float64(cMaximumSwapFlag.Name),
		MinimumSwap:       ctx.Float64(cMinimumSwapFlag.Name),
		BigValueThreshold: ctx.Float64(cBigValueThresholdFlag.Name),
		SwapFeeRate:       ctx.Float64(cSwapFeeRateFlag.Name),
		MaximumSwapFee:    ctx.Float64(cMaximumSwapFeeFlag.Name),
		MinimumSwapFee:    ctx.Float64(cMinimumSwapFeeFlag.Name),
	}
	err = tokenCfg.CheckConfig()
	if err != nil {
		return err
	}
	jsdata, err := json.MarshalIndent(tokenCfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("chainID is", chainID)
	fmt.Println("token config struct is", string(jsdata))
	funcHash := common.FromHex("0xbb14e9ff")
	inputData := router.PackDataWithFuncHash(funcHash,
		common.BytesToHash([]byte(tokenCfg.TokenID)),
		chainID,
		decimals,
		common.HexToAddress(tokenCfg.ContractAddress),
		tokenCfg.ContractVersion,
		router.ToBits(tokenCfg.MaximumSwap, decimals),
		router.ToBits(tokenCfg.MinimumSwap, decimals),
		router.ToBits(tokenCfg.BigValueThreshold, decimals),
		uint64(tokenCfg.SwapFeeRate*1000000),
		router.ToBits(tokenCfg.MaximumSwapFee, decimals),
		router.ToBits(tokenCfg.MinimumSwapFee, decimals),
	)
	fmt.Println("set token config input data is", common.ToHex(inputData))
	return nil
}
