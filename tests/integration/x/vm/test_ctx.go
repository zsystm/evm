package vm

import (
	"github.com/cosmos/evm/testutil/integration/evm/network"
	evmante "github.com/cosmos/evm/x/vm/ante"

	storetypes "cosmossdk.io/store/types"
)

func (suite *EvmAnteTestSuite) TestBuildEvmExecutionCtx() {
	network := network.New(suite.create, nil)

	ctx := evmante.BuildEvmExecutionCtx(network.GetContext())

	suite.Equal(storetypes.GasConfig{}, ctx.KVGasConfig())
	suite.Equal(storetypes.GasConfig{}, ctx.TransientKVGasConfig())
}
