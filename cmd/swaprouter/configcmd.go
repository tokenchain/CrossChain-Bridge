package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

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
				Name:   "genChainConfigData",
				Usage:  "generate ChainConfig json marshal data",
				Action: genChainConfigData,
				Flags: []cli.Flag{
					cBlockChainFlag,
					cChainIDFlag,
					cConfirmationsFlag,
					cRouterContractFlag,
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
				Name:   "genTokenConfigData",
				Usage:  "generate TokenConfig json marshal data",
				Action: genTokenConfigData,
				Flags: []cli.Flag{
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
			{
				Name:      "packCallStrArray",
				Usage:     "pack data calling method with one string array",
				ArgsUsage: "<methodName> [args...]",
				Action:    packCallStrArray,
				Flags:     []cli.Flag{},
				Description: `
pack input data of calling method with just one string array parameter
(eg. setTokenIDs(string[])) in the way same as abi coder v2.
`,
			},
			{
				Name:      "decodeHex",
				Usage:     "decode hex string",
				ArgsUsage: "<hex string>",
				Action:    decodeHex,
				Flags:     []cli.Flag{},
				Description: `
decode hex string
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

	cContractVersionFlag = &cli.Float64Flag{
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

func genChainConfigData(ctx *cli.Context) error {
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
	jsdata, err = json.Marshal(chainCfg)
	if err != nil {
		return err
	}
	fmt.Println("chain config data is", common.ToHex(jsdata))
	return nil
}

func genTokenConfigData(ctx *cli.Context) error {
	decimals := ctx.Int(cDecimalsFlag.Name)
	if decimals < 0 || decimals > 256 {
		return fmt.Errorf("wrong decimals '%v'", decimals)
	}
	tokenCfg := &router.TokenConfig{
		TokenID:           ctx.String(cTokenIDFlag.Name),
		Decimals:          uint8(decimals),
		ContractAddress:   ctx.String(cContractAddressFlag.Name),
		ContractVersion:   ctx.Float64(cContractVersionFlag.Name),
		MaximumSwap:       ctx.Float64(cMaximumSwapFlag.Name),
		MinimumSwap:       ctx.Float64(cMinimumSwapFlag.Name),
		BigValueThreshold: ctx.Float64(cBigValueThresholdFlag.Name),
		SwapFeeRate:       ctx.Float64(cSwapFeeRateFlag.Name),
		MaximumSwapFee:    ctx.Float64(cMaximumSwapFeeFlag.Name),
		MinimumSwapFee:    ctx.Float64(cMinimumSwapFeeFlag.Name),
	}
	err := tokenCfg.CheckConfig()
	if err != nil {
		return err
	}
	jsdata, err := json.MarshalIndent(tokenCfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("token config struct is", string(jsdata))
	jsdata, err = json.Marshal(tokenCfg)
	if err != nil {
		return err
	}
	fmt.Println("token config data is", common.ToHex(jsdata))
	return nil
}

func decodeHex(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		_ = cli.ShowCommandHelp(ctx, "decodeHex")
		fmt.Println()
		return fmt.Errorf("miss argument of hex string")
	}
	hexStr := ctx.Args().Get(0)
	if common.HasHexPrefix(hexStr) {
		hexStr = hexStr[2:]
	}
	hexData, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	fmt.Println(string(hexData))
	return nil
}

func packCallStrArray(ctx *cli.Context) error {
	argsCount := ctx.NArg()
	if argsCount < 1 {
		_ = cli.ShowCommandHelp(ctx, "packCallStrArray")
		fmt.Println()
		return fmt.Errorf("miss argument of method name and args")
	}
	methodName := ctx.Args().Get(0)
	args := make([]string, argsCount-1)
	for i := 1; i < argsCount; i++ {
		args[i-1] = ctx.Args().Get(i)
	}
	funcProto := fmt.Sprintf("%v(string[])", methodName)
	funcHash := common.Keccak256Hash([]byte(funcProto))
	input := router.PackDataWithFuncHash(funcHash[:4], args)
	fmt.Println(common.ToHex(input))
	return nil
}
