package ante

import (
	"cosmosevm.io/evmd/tests/integration"
	"testing"

	"github.com/cosmos/evm/tests/integration/ante"
)

func TestAnte_Integration(t *testing.T) {
	ante.TestIntegrationAnteHandler(t, integration.CreateEvmd)
}
