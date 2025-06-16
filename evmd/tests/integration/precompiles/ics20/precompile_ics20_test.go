package ics20

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/evmd/tests/integration"
	"github.com/cosmos/evm/tests/integration/precompiles/ics20"
)

func TestGovPrecompileTestSuite(t *testing.T) {
	s := ics20.NewPrecompileTestSuite(t, integration.SetupEvmd)
	suite.Run(t, s)
}

func TestGovPrecompileIntegrationTestSuite(t *testing.T) {
	ics20.TestPrecompileIntegrationTestSuite(t, integration.SetupEvmd)
}
