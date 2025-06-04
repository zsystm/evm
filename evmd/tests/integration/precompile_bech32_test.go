package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/bech32"
)

func TestBech32PrecompileTestSuite(t *testing.T) {
	s := bech32.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}
