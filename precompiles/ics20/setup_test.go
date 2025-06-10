package ics20_test

import (
	"github.com/cosmos/evm/evmd"
	evmibctesting "github.com/cosmos/evm/ibc/testing"
	"github.com/cosmos/evm/precompiles/ics20"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PrecompileTestSuite struct {
	suite.Suite

	coordinator *evmibctesting.Coordinator

	chainA           *evmibctesting.TestChain
	chainAPrecompile *ics20.Precompile
	chainB           *evmibctesting.TestChain
	chainBPrecompile *ics20.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	s.coordinator = evmibctesting.NewCoordinator(s.T(), 2, 0)
	s.chainA = s.coordinator.GetChain(evmibctesting.GetEvmChainID(1))
	s.chainB = s.coordinator.GetChain(evmibctesting.GetEvmChainID(2))

	evmAppA := s.chainA.App.(*evmd.EVMD)
	s.chainAPrecompile, _ = ics20.NewPrecompile(
		*evmAppA.StakingKeeper,
		evmAppA.TransferKeeper,
		evmAppA.IBCKeeper.ChannelKeeper,
		evmAppA.EVMKeeper,
	)
	evmAppB := s.chainB.App.(*evmd.EVMD)
	s.chainBPrecompile, _ = ics20.NewPrecompile(
		*evmAppB.StakingKeeper,
		evmAppB.TransferKeeper,
		evmAppB.IBCKeeper.ChannelKeeper,
		evmAppB.EVMKeeper,
	)
}
