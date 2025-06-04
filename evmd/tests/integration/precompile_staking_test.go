package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/staking"
)

func TestStakingPrecompileTestSuite(t *testing.T) {
	s := staking.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestStakingPrecompileIntegrationTestSuite(t *testing.T) {
	staking.TestPrecompileIntegrationTestSuite(t, CreateEvmd)
}
