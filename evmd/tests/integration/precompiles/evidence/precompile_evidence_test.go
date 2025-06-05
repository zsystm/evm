package evidence

import (
	"cosmosevm.io/evmd/tests/integration"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/evidence"
)

func TestEvidencePrecompileTestSuite(t *testing.T) {
	s := evidence.NewPrecompileTestSuite(integration.CreateEvmd)
	suite.Run(t, s)
}
