package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	examplechain "github.com/cosmos/evm/example_chain"
	"github.com/cosmos/evm/example_chain/evmd/cmd"
	chainconfig "github.com/cosmos/evm/example_chain/evmd/config"
)

func main() {
	setupSDKConfig()

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "evmd", examplechain.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func setupSDKConfig() {
	config := sdk.GetConfig()
	chainconfig.SetBech32Prefixes(config)
	config.Seal()
}
