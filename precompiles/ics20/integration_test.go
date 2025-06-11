package ics20_test

import (
	"fmt"
	"github.com/cosmos/evm/evmd"
	"github.com/cosmos/evm/precompiles/testutil/contracts"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	var _ = Describe("Calling ICS20 precompile from contract", func() {
		s := new(PrecompileTestSuite)

		var (
			ics20CallerContract evmtypes.CompiledContract
			//contractAddr        common.Address
			//callArgs            factory.CallArgs
			//txArgs              evmtypes.EvmTxArgs
			//defaultLogCheck     testutil.LogCheckArgs
			//execRevertedCheck   testutil.LogCheckArgs
			//randomAccAddr       sdk.AccAddress
			err error
		)

		BeforeEach(func() {
			ics20CallerContract, err = contracts.LoadIcs20CallerContract()
			Expect(err).To(BeNil())

			s.internalT = t
			s.SetupTest()

			sender := s.chainA.SenderAccount.GetAddress()
			res, sentEthTx, _, err := s.chainA.SendEvmTx(
				s.chainA.SenderPrivKey,
				common.Address{},
				big.NewInt(0),
				ics20CallerContract.Bin,
			)
			Expect(err).To(BeNil())
			Expect(res.Code).To(BeZero(), "Failed to deploy ICS20 caller contract: %s", res.Log)

			contractAddr := crypto.CreateAddress(common.Address(sender.Bytes()), sentEthTx.AsTransaction().Nonce())
			evmAppA := s.chainA.App.(*evmd.EVMD)
			ctxA := s.chainA.GetContext()
			codeHash := evmAppA.EVMKeeper.GetCodeHash(ctxA, contractAddr)
			code := evmAppA.EVMKeeper.GetCode(ctxA, codeHash)
			Expect(code).To(Equal(ics20CallerContract.Bin), "Deployed contract code does not match expected code")

			//randomAccAddr = sdk.AccAddress(tx.GenerateAddress().Bytes())
		})

		It("test", func() {
			fmt.Println("Random")
		})

		//It("should fail if sender is different from msg.sender", func() {
		//	portID := transfertypes.PortID
		//	channelID := "channel-0"
		//	connectionID := "connection-0"
		//	clientID := "07-tendermint"
		//	version := transfertypes.V1
		//	var connectionVersion *connectiontypes.Version
		//
		//	txSender := s.keyring.GetAccAddr(0)
		//	txSenderKey := s.keyring.GetPrivKey(0)
		//	// Set up the connection
		//	msgConnectionOpenInit := connectiontypes.NewMsgConnectionOpenInit(
		//		clientID, clientID,
		//		commitmenttypes.NewMerklePrefix([]byte("storePrefixKey")),
		//		connectionVersion, 500, txSender.String(),
		//	)
		//	resp, err := s.factory.ExecuteCosmosTx(txSenderKey, commonfactory.CosmosTxArgs{
		//		Msgs: []sdk.Msg{msgConnectionOpenInit},
		//	})
		//	Expect(err).To(BeNil())
		//	Expect(resp.Code).To(BeZero(), "Failed to open channel: %s", resp.Log)
		//
		//	// Set up the channel
		//	msgChannelOpenInit := channeltypes.NewMsgChannelOpenInit(
		//		portID, version, channeltypes.UNORDERED,
		//		[]string{connectionID}, portID, txSender.String(),
		//	)
		//	resp, err = s.factory.ExecuteCosmosTx(txSenderKey, commonfactory.CosmosTxArgs{
		//		Msgs: []sdk.Msg{msgChannelOpenInit},
		//	})
		//	Expect(err).To(BeNil())
		//	Expect(resp.Code).To(BeZero(), "Failed to open channel: %s", resp.Log)
		//
		//	anotherAddr := s.keyring.GetAddr(1)
		//	callArgs.MethodName = "testIbcTransfer"
		//	callArgs.Args = []interface{}{
		//		portID,
		//		channelID,
		//		s.bondDenom,
		//		big.NewInt(1),
		//		anotherAddr,
		//		randomAccAddr.String(),
		//		ics20.DefaultTimeoutHeight,
		//		uint64(time.Now().Add(time.Minute).Unix()),
		//		"",
		//	}
		//
		//	check := defaultLogCheck.WithErrContains(
		//		cmn.ErrRequesterIsNotMsgSender, contractAddr.String(), s.keyring.GetAddr(0).String(),
		//	)
		//	_, _, err = s.factory.CallContractAndCheckLogs(txSenderKey, txArgs, callArgs, check)
		//	Expect(err).To(BeNil())
		//})
		//
		//It("should fail if channel does not exist", func() {
		//	callArgs.MethodName = "testIbcTransferFromContract"
		//	callArgs.Args = []interface{}{
		//		transfertypes.PortID,
		//		"channel-0",
		//		s.bondDenom,
		//		big.NewInt(1),
		//		s.keyring.GetAddr(0).String(),
		//		ics20.DefaultTimeoutHeight,
		//		uint64(time.Now().Add(time.Minute).Unix()),
		//		"",
		//	}
		//
		//	_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, execRevertedCheck)
		//	Expect(err).To(BeNil())
		//})
		//
	})

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}
