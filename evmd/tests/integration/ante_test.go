package integration

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/ante"
)

func TestAnte_Integration(t *testing.T) {
	ante.TestIntegrationAnteHandler(t, CreateEvmd)
}
