package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/gov"
)

func TestGovPrecompileTestSuite(t *testing.T) {
	s := gov.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestGovPrecompileIntegrationTestSuite(t *testing.T) {
	gov.TestPrecompileIntegrationTestSuite(t, CreateEvmd)
}
