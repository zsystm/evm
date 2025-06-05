package eips

import (
	"cosmosevm.io/evmd/tests/integration"
	"testing"

	"github.com/cosmos/evm/tests/integration/eips"
)

func Test_EIPs(t *testing.T) {
	eips.TestEIPs(t, integration.CreateEvmd)
}
