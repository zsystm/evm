package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/x/vm"
)

func TestKeeperTestSuite(t *testing.T) {
	s := vm.NewKeeperTestSuite(CreateEvmd)
	s.EnableFeemarket = false
	s.EnableLondonHF = true
	suite.Run(t, s)
}
