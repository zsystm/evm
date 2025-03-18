package constants_test

import (
	"testing"

	"github.com/cosmos/evm/testutil/constants"

	"github.com/cosmos/evm/example_chain"
	chainconfig "github.com/cosmos/evm/example_chain/evmd/config"
	"github.com/stretchr/testify/require"
)

func TestRequireSameTestDenom(t *testing.T) {
	require.Equal(t,
		constants.ExampleAttoDenom,
		example_chain.ExampleChainDenom,
		"test denoms should be the same across the repo",
	)
}

func TestRequireSameTestBech32Prefix(t *testing.T) {
	require.Equal(t,
		constants.ExampleBech32Prefix,
		chainconfig.Bech32Prefix,
		"bech32 prefixes should be the same across the repo",
	)
}

func TestRequireSameWEVMOSMainnet(t *testing.T) {
	require.Equal(t,
		constants.WEVMOSContractMainnet,
		example_chain.WEVMOSContractMainnet,
		"wevmos contract addresses should be the same across the repo",
	)
}
