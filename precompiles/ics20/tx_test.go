package ics20_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/evmd"
	evmibctesting "github.com/cosmos/evm/ibc/testing"
	"github.com/cosmos/evm/testutil/tx"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type testCase struct {
	name               string
	port               string
	channelID          string
	useDynamicChannel  bool
	overrideSender     bool
	receiver           string
	expectErrSubstring string
}

func (s *PrecompileTestSuite) TestTransferErrors() {
	evmAppA := s.chainA.App.(*evmd.EVMD)
	denom, err := evmAppA.StakingKeeper.BondDenom(s.chainA.GetContext())
	s.Require().NoError(err)

	timeoutHeight := clienttypes.NewHeight(1, 110)
	amount := sdkmath.NewInt(1)
	defaultSender := common.BytesToAddress(s.chainA.SenderAccount.GetAddress().Bytes())
	defaultReceiver := s.chainB.SenderAccount.GetAddress().String()

	tests := []testCase{
		{
			name:               "invalid source channel",
			port:               transfertypes.PortID,
			channelID:          "invalid/channel",
			receiver:           defaultReceiver,
			expectErrSubstring: "invalid source channel ID",
		},
		{
			name:               "channel not found",
			port:               transfertypes.PortID,
			channelID:          "channel-9",
			receiver:           defaultReceiver,
			expectErrSubstring: "channel not found",
		},
		{
			name:               "invalid receiver",
			port:               transfertypes.PortID,
			useDynamicChannel:  true,
			receiver:           "",
			expectErrSubstring: "invalid address",
		},
		{
			name:               "msg sender is not a contract caller",
			port:               transfertypes.PortID,
			useDynamicChannel:  true,
			overrideSender:     true,
			receiver:           defaultReceiver,
			expectErrSubstring: "does not match the requester address",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			path := evmibctesting.NewTransferPath(s.chainA, s.chainB)
			path.Setup()

			channel := tc.channelID
			if tc.useDynamicChannel {
				channel = path.EndpointA.ChannelID
			}

			sender := defaultSender
			if tc.overrideSender {
				sender = tx.GenerateAddress()
			}

			data, err := s.chainAPrecompile.ABI.Pack(
				"transfer",
				tc.port,
				channel,
				denom,
				amount.BigInt(),
				sender,
				tc.receiver,
				timeoutHeight,
				uint64(0),
				"",
			)
			s.Require().NoError(err)

			_, err = s.chainA.SendEvmTx(
				s.chainA.SenderPrivKey,
				s.chainAPrecompile.Address(),
				big.NewInt(0),
				data,
			)
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.expectErrSubstring)
		})
	}
}

func (s *PrecompileTestSuite) TestTransfer() {
	path := evmibctesting.NewTransferPath(s.chainA, s.chainB)
	path.Setup()

	evmAppA := s.chainA.App.(*evmd.EVMD)
	denom, err := evmAppA.StakingKeeper.BondDenom(s.chainA.GetContext())
	s.Require().NoError(err)

	amount := sdkmath.NewInt(5)
	sourceAddr := common.BytesToAddress(s.chainA.SenderAccount.GetAddress().Bytes())
	receiver := s.chainB.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(1, 110)

	sourcePort := path.EndpointA.ChannelConfig.PortID
	sourceChannel := path.EndpointA.ChannelID
	data, err := s.chainAPrecompile.ABI.Pack(
		"transfer",
		sourcePort,
		sourceChannel,
		denom,
		amount.BigInt(),
		sourceAddr,
		receiver,
		timeoutHeight,
		uint64(0),
		"",
	)
	s.Require().NoError(err)

	res, err := s.chainA.SendEvmTx(
		s.chainA.SenderPrivKey,
		s.chainAPrecompile.Address(),
		big.NewInt(0),
		data,
	)
	s.Require().NoError(err)

	packet, err := evmibctesting.ParsePacketFromEvents(res.Events)
	s.Require().NoError(err)

	err = path.RelayPacket(packet)
	s.Require().NoError(err)

	trace := transfertypes.NewHop(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
	chainBDenom := transfertypes.NewDenom(denom, trace)
	evmAppB := s.chainB.App.(*evmd.EVMD)
	balance := evmAppB.BankKeeper.GetBalance(
		s.chainB.GetContext(),
		s.chainB.SenderAccount.GetAddress(),
		chainBDenom.IBCDenom(),
	)
	s.Require().Equal(sdk.NewCoin(chainBDenom.IBCDenom(), amount), balance)
}
