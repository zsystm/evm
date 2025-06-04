package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/p256"
)

func TestP256PrecompileTestSuite(t *testing.T) {
	s := p256.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestP256PrecompileIntegrationTestSuite(t *testing.T) {
	p256.TestPrecompileIntegrationTestSuite(t, CreateEvmd)
}
