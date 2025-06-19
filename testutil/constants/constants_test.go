package constants_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	evmconfig "github.com/cosmos/evm/config"
	"github.com/cosmos/evm/testutil/constants"
)

func TestRequireSameTestDenom(t *testing.T) {
	require.Equal(t,
		constants.ExampleAttoDenom,
		evmconfig.ExampleChainDenom,
		"test denoms should be the same across the repo",
	)
}

func TestRequireSameTestBech32Prefix(t *testing.T) {
	require.Equal(t,
		constants.ExampleBech32Prefix,
		evmconfig.Bech32Prefix,
		"bech32 prefixes should be the same across the repo",
	)
}

func TestRequireSameWEVMOSMainnet(t *testing.T) {
	require.Equal(t,
		constants.WEVMOSContractMainnet,
		evmconfig.WEVMOSContractMainnet,
		"wevmos contract addresses should be the same across the repo",
	)
}
