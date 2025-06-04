package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	erc21 "github.com/cosmos/evm/tests/integration/precompiles/erc20"
)

func TestErc20PrecompileTestSuite(t *testing.T) {
	s := erc21.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestErc20IntegrationTestSuite(t *testing.T) {
	erc21.TestIntegrationTestSuite(t, CreateEvmd)
}
