package ics20_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/evmd"
	evmibctesting "github.com/cosmos/evm/ibc/testing"
	"github.com/cosmos/evm/precompiles/ics20"
	"github.com/cosmos/evm/precompiles/testutil/contracts"
	"github.com/cosmos/evm/testutil/integration/os/factory"
	"github.com/cosmos/evm/testutil/tx"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	isContractDeployed := func(ctx sdk.Context, evmApp *evmd.EVMD, contractAddr common.Address) bool {
		codeHash := evmApp.EVMKeeper.GetCodeHash(ctx, contractAddr)
		code := evmApp.EVMKeeper.GetCode(ctx, codeHash)
		return len(code) > 0
	}

	var _ = Describe("Calling ICS20 precompile from contract", func() {
		s := new(PrecompileTestSuite)

		var (
			ics20CallerContract evmtypes.CompiledContract
			ics20CallerAddr     common.Address
			//defaultLogCheck     testutil.LogCheckArgs
			//execRevertedCheck   testutil.LogCheckArgs
			randomAddr    common.Address
			randomAccAddr sdk.AccAddress
			err           error
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

			ics20CallerAddr = crypto.CreateAddress(common.Address(sender.Bytes()), sentEthTx.AsTransaction().Nonce())
			evmAppA := s.chainA.App.(*evmd.EVMD)
			Expect(isContractDeployed(s.chainA.GetContext(), evmAppA, ics20CallerAddr)).To(BeTrue(), "Contract was not deployed successfully")

			randomAddr = tx.GenerateAddress()
			randomAccAddr = sdk.AccAddress(randomAddr.Bytes())
		})

		It("should fail if send is different from msg.sender (only direct call is allowed, not for proxy)", func() {
			path := evmibctesting.NewTransferPath(s.chainA, s.chainB)
			path.Setup()

			sourcePortID := path.EndpointA.ChannelConfig.PortID
			sourceChannelID := path.EndpointA.ChannelID
			sourceBondDenom := s.chainABondDenom
			sender := common.Address(s.chainA.SenderAccount.GetAddress().Bytes())

			callArgs := factory.CallArgs{
				ContractABI: ics20CallerContract.ABI,
				MethodName:  "testIbcTransfer",
				Args: []interface{}{
					sourcePortID,
					sourceChannelID,
					sourceBondDenom,
					big.NewInt(1),
					sender,
					randomAccAddr.String(),
					ics20.DefaultTimeoutHeight,
					uint64(time.Now().Add(time.Minute).Unix()),
					"",
				},
			}
			input, err := factory.GenerateContractCallArgsInput(callArgs)
			Expect(err).To(BeNil(), "Failed to generate contract call args")
			_, _, _, err = s.chainA.SendEvmTx(
				s.chainA.SenderPrivKey,
				ics20CallerAddr,
				big.NewInt(0),
				input,
			)
			Expect(err).NotTo(BeNil(), "Failed to testTransfer: %s", err.Error())
		})

		It("should fail if the v1 channel is not found", func() {
			path := evmibctesting.NewTransferPath(s.chainA, s.chainB)
			path.Setup()

			sourcePortID := path.EndpointA.ChannelConfig.PortID
			nonExistentChannelID := "channel-100"
			sourceBondDenom := s.chainABondDenom

			callArgs := factory.CallArgs{
				ContractABI: ics20CallerContract.ABI,
				MethodName:  "testIbcTransfer",
				Args: []interface{}{
					sourcePortID,
					nonExistentChannelID,
					sourceBondDenom,
					big.NewInt(1),
					ics20CallerAddr,
					randomAccAddr.String(),
					ics20.DefaultTimeoutHeight,
					uint64(time.Now().Add(time.Minute).Unix()),
					"",
				},
			}
			input, err := factory.GenerateContractCallArgsInput(callArgs)
			Expect(err).To(BeNil(), "Failed to generate contract call args")
			_, _, _, err = s.chainA.SendEvmTx(
				s.chainA.SenderPrivKey,
				ics20CallerAddr,
				big.NewInt(0),
				input,
			)
			Expect(err).NotTo(BeNil(), "Failed to testTransfer: %s", err.Error())
		})

		It("should fail if the v2 client id format is invalid", func() {
			path := evmibctesting.NewTransferPath(s.chainA, s.chainB)
			path.Setup()

			sourcePortID := path.EndpointA.ChannelConfig.PortID
			invalidV2ClientID := "v2"
			sourceBondDenom := s.chainABondDenom

			callArgs := factory.CallArgs{
				ContractABI: ics20CallerContract.ABI,
				MethodName:  "testIbcTransfer",
				Args: []interface{}{
					sourcePortID,
					invalidV2ClientID,
					sourceBondDenom,
					big.NewInt(1),
					ics20CallerAddr,
					randomAccAddr.String(),
					ics20.DefaultTimeoutHeight,
					uint64(time.Now().Add(time.Minute).Unix()),
					"",
				},
			}
			input, err := factory.GenerateContractCallArgsInput(callArgs)
			Expect(err).To(BeNil(), "Failed to generate contract call args")
			_, _, _, err = s.chainA.SendEvmTx(
				s.chainA.SenderPrivKey,
				ics20CallerAddr,
				big.NewInt(0),
				input,
			)
			Expect(err).NotTo(BeNil(), "Failed to testTransfer: %s", err.Error())
		})
	})

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}
