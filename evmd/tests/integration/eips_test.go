package integration

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/eips"
)

func Test_EIPs(t *testing.T) {
	eips.TestEIPs(t, CreateEvmd)
}
