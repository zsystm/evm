package ics20_test

import (
	"github.com/cosmos/evm/evmd"
	evmibctesting "github.com/cosmos/evm/ibc/testing"
	"github.com/cosmos/evm/precompiles/ics20"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"
	"github.com/cosmos/evm/testutil/integration/os/network"
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
	// Setup IBC
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

type IntegrationTestSuite struct {
	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	bondDenom  string
	precompile *ics20.Precompile
}

func (s *IntegrationTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	ctx := nw.GetContext()
	sk := nw.App.StakingKeeper
	denom, err := sk.BondDenom(ctx)
	if err != nil {
		panic(err)
	}

	s.network = nw
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.bondDenom = denom

	s.precompile, err = ics20.NewPrecompile(
		*nw.App.StakingKeeper,
		nw.App.TransferKeeper,
		nw.App.IBCKeeper.ChannelKeeper,
		nw.App.EVMKeeper,
	)
	if err != nil {
		panic(err)
	}
}
