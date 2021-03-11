package main

import (
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/admin"
	"github.com/anyswap/CrossChain-Bridge/cmd/utils"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/urfave/cli/v2"
)

var (
	addpairCommand = &cli.Command{
		Action:    addpair,
		Name:      "addpair",
		Usage:     "add token pair",
		ArgsUsage: "<configFile>",
		Description: `
add token pair dynamically through config file
`,
		Flags: admin.CommonFlags,
	}
)

func addpair(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	method := "addpair"
	if ctx.NArg() != 1 {
		_ = cli.ShowCommandHelp(ctx, method)
		fmt.Println()
		return fmt.Errorf("invalid arguments: %q", ctx.Args())
	}

	err := admin.Prepare(ctx)
	if err != nil {
		return err
	}

	configFile := ctx.Args().Get(0)

	log.Printf("admin addpair: %v", configFile)

	params := []string{configFile}
	result, err := admin.SwapAdmin(method, params)

	log.Printf("result is '%v'", result)
	return err
}
