package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/evidence"
)

func TestEvidencePrecompileTestSuite(t *testing.T) {
	s := evidence.NewPrecompileTestSuite(CreateEvmd)
	suite.Run(t, s)
}
