package ics20_test

import (
	"cosmossdk.io/math"
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/ics20"
	"github.com/cosmos/evm/precompiles/testutil"
	"github.com/cosmos/evm/precompiles/testutil/contracts"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	testutils "github.com/cosmos/evm/testutil/integration/os/utils"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
	"testing"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}

//var _ = Describe("Calling ICS20 Precompile from EOA", func() {
//	var s *IntegrationTestSuite
//	var path *evmibctesting.Path
//
//	BeforeEach(func() {
//		s = new(PrecompileTestSuite)
//		s.SetupTest()
//
//		path = evmibctesting.NewPath(s.chainA, s.chainB)
//		path.Setup()
//	})
//
//	Describe("Execute ICS20 Transfer", func() {
//		//var sourcePort, sourceChannel, receiver string
//		BeforeEach(func() {
//			//sourcePort = path.EndpointA.ChannelConfig.PortID
//			//sourceChannel = path.EndpointA.ChannelID
//			//receiver = s.chainB.SenderAccount.GetAddress().String()
//
//			ics20CallerContract, err := contracts.LoadIcs20CallerContract()
//			Expect(err).ToNot(HaveOccurred(), "failed to load ICS20 caller contract")
//
//			// Deploy the ics20 caller contract
//			data := ics20CallerContract.Bin
//			res, err := s.chainA.SendEvmTx(
//				s.chainA.SenderPrivKey,
//				s.chainAPrecompile.Address(),
//				big.NewInt(0),
//				data,
//			)
//			Expect(err).ToNot(HaveOccurred(), "failed to deploy ICS20 caller contract")
//			Expect(res).ToNot(BeNil(), "ICS20 caller contract deployment response should not be nil")
//		})
//
//		It("should successfully transfer tokens", func() {
//			fmt.Println("DUMMY TEST")
//		})
//
//	})
//})

var _ = Describe("Calling ICS20 precompile from contract", func() {
	s := new(IntegrationTestSuite)

	var (
		ics20CallerContract evmtypes.CompiledContract
		contractAddr        common.Address
		callArgs            factory.CallArgs
		txArgs              evmtypes.EvmTxArgs
		defaultLogCheck     testutil.LogCheckArgs
		execRevertedCheck   testutil.LogCheckArgs
		err                 error
	)

	BeforeEach(func() {
		ics20CallerContract, err = contracts.LoadIcs20CallerContract()
		Expect(err).To(BeNil())

		s.SetupTest()

		contractAddr, err = s.factory.DeployContract(
			s.keyring.GetPrivKey(0),
			evmtypes.EvmTxArgs{},
			factory.ContractDeploymentData{Contract: ics20CallerContract},
		)
		Expect(err).To(BeNil())
		Expect(s.network.NextBlock()).To(BeNil())

		err = testutils.FundAccountWithBaseDenom(s.factory, s.network, s.keyring.GetKey(0), contractAddr.Bytes(), math.NewInt(1e18))
		Expect(err).To(BeNil())
		Expect(s.network.NextBlock()).To(BeNil())

		callArgs = factory.CallArgs{ContractABI: ics20CallerContract.ABI}
		txArgs = evmtypes.EvmTxArgs{To: &contractAddr}
		defaultLogCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.ABI.Events}
		execRevertedCheck = defaultLogCheck.WithErrContains(vm.ErrExecutionReverted.Error())
	})

	It("should fail if sender is different from msg.sender", func() {
		callArgs.MethodName = "testIbcTransfer"
		callArgs.Args = []interface{}{
			transfertypes.PortID,
			"channel-0",
			s.bondDenom,
			big.NewInt(1),
			s.keyring.GetAddr(0),
			s.keyring.GetAddr(0).String(),
			ics20.DefaultTimeoutHeight,
			uint64(time.Now().Add(time.Minute).Unix()),
			"",
		}

		check := defaultLogCheck.WithErrContains(
			cmn.ErrRequesterIsNotMsgSender, contractAddr.String(), s.keyring.GetAddr(0).String(),
		)
		_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, check)
		Expect(err).To(BeNil())
	})

	It("should fail if channel does not exist", func() {
		callArgs.MethodName = "testIbcTransferFromContract"
		callArgs.Args = []interface{}{
			transfertypes.PortID,
			"channel-0",
			s.bondDenom,
			big.NewInt(1),
			s.keyring.GetAddr(0).String(),
			ics20.DefaultTimeoutHeight,
			uint64(time.Now().Add(time.Minute).Unix()),
			"",
		}

		_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, execRevertedCheck)
		Expect(err).To(BeNil())
	})
})
