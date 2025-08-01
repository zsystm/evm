package p256

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/cometbft/cometbft/crypto"

	"github.com/cosmos/evm/precompiles/p256"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	"github.com/cosmos/evm/testutil/integration/evm/network"
	"github.com/cosmos/evm/testutil/integration/evm/utils"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

type IntegrationTestSuite struct {
	network           network.Network
	factory           factory.TxFactory
	keyring           testkeyring.Keyring
	precompileAddress common.Address
	p256Priv          *ecdsa.PrivateKey
}

func TestPrecompileIntegrationTestSuite(t *testing.T, create network.CreateEvmApp, options ...network.ConfigOption) {
	_ = Describe("Calling p256 precompile directly", Label("P256 Precompile"), Ordered, func() {
		var s *IntegrationTestSuite

		BeforeAll(func() {
			keyring := testkeyring.New(1)
			opts := []network.ConfigOption{
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			}
			opts = append(opts, options...)
			integrationNetwork := network.New(create, opts...)
			grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
			txFactory := factory.New(integrationNetwork, grpcHandler)
			p256Priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			Expect(err).To(BeNil())

			s = &IntegrationTestSuite{
				network:           integrationNetwork,
				factory:           txFactory,
				keyring:           keyring,
				precompileAddress: p256.Precompile{}.Address(),
				p256Priv:          p256Priv,
			}
		})

		AfterEach(func() {
			// Start each test with a fresh block
			err := s.network.NextBlock()
			Expect(err).To(BeNil())
		})

		When("the precompile is enabled in the EVM params", func() {
			BeforeAll(func() {
				s = setupIntegrationTestSuite(nil, create, options...)
			})

			DescribeTable("execute contract call", func(inputFn func() (input, expOutput []byte, expErr string)) {
				senderKey := s.keyring.GetKey(0)

				input, expOutput, expErr := inputFn()
				args := evmtypes.EvmTxArgs{
					To:    &s.precompileAddress,
					Input: input,
				}

				txResult, err := s.factory.ExecuteEthTx(senderKey.Priv, args)
				Expect(err).To(BeNil())
				Expect(txResult.IsOK()).To(Equal(true), "transaction should have succeeded", txResult.GetLog())

				res, err := utils.DecodeExecTxResult(txResult)
				Expect(err).To(BeNil())
				Expect(res.VmError).To(Equal(expErr), "expected different vm error")
				Expect(res.Ret).To(Equal(expOutput))
			},
				Entry(
					"valid signature",
					func() (input, expOutput []byte, expErr string) {
						input, err := signMsg([]byte("hello world"), s.p256Priv)
						Expect(err).To(BeNil())
						return input, trueValue, ""
					},
				),
				Entry(
					"invalid signature",
					func() (input, expOutput []byte, expErr string) {
						privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
						Expect(err).To(BeNil())

						hash := crypto.Sha256([]byte("hello world"))

						rInt, sInt, err := ecdsa.Sign(rand.Reader, s.p256Priv, hash)
						Expect(err).To(BeNil())
						pub := privB.PublicKey

						input = make([]byte, p256.VerifyInputLength)
						copy(input[0:32], hash)

						// ALWAYS left-pad to 32 bytes:
						rInt.FillBytes(input[32:64])
						sInt.FillBytes(input[64:96])
						pub.X.FillBytes(input[96:128])
						pub.Y.FillBytes(input[128:160])
						return input, nil, ""
					},
				),
			)
		})

		When("the precompile is not enabled in the EVM params", func() {
			BeforeAll(func() {
				customGenesis := evmtypes.DefaultGenesisState()
				customGenesis.Params.ActiveStaticPrecompiles = evmtypes.AvailableStaticPrecompiles
				params := customGenesis.Params
				addr := s.precompileAddress.String()
				var activePrecompiles []string
				for _, precompile := range params.ActiveStaticPrecompiles {
					if precompile != addr {
						activePrecompiles = append(activePrecompiles, precompile)
					}
				}
				params.ActiveStaticPrecompiles = activePrecompiles
				customGenesis.Params = params
				s = setupIntegrationTestSuite(customGenesis, create, options...)
			})

			DescribeTable("execute contract call", func(inputFn func() (input []byte)) {
				senderKey := s.keyring.GetKey(0)

				input := inputFn()
				args := evmtypes.EvmTxArgs{
					To:    &s.precompileAddress,
					Input: input,
				}

				_, err := s.factory.ExecuteEthTx(senderKey.Priv, args)
				Expect(err).To(BeNil(), "expected no error since contract doesn't exists")
			},
				Entry(
					"valid signature",
					func() (input []byte) {
						input, err := signMsg([]byte("hello world"), s.p256Priv)
						Expect(err).To(BeNil())
						return input
					},
				),
				Entry(
					"invalid signature",
					func() (input []byte) {
						privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
						Expect(err).To(BeNil())

						hash := crypto.Sha256([]byte("hello world"))

						rInt, sInt, err := ecdsa.Sign(rand.Reader, s.p256Priv, hash)
						Expect(err).To(BeNil())

						input = make([]byte, p256.VerifyInputLength)
						copy(input[0:32], hash)
						rInt.FillBytes(input[32:64])
						sInt.FillBytes(input[64:96])
						privB.PublicKey.X.FillBytes(input[96:128])
						privB.PublicKey.Y.FillBytes(input[128:160])
						return input
					},
				),
			)
		})
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "P256 Precompile Integration Test Suite")
}

// setupIntegrationTestSuite is a helper function to setup a integration test suite
// with a network with a specified custom genesis state for the EVM module
func setupIntegrationTestSuite(customEVMGenesis *evmtypes.GenesisState, create network.CreateEvmApp, options ...network.ConfigOption) *IntegrationTestSuite {
	customGenesis := network.CustomGenesisState{}
	if customEVMGenesis != nil {
		customGenesis[evmtypes.ModuleName] = customEVMGenesis
	}
	keyring := testkeyring.New(1)
	opts := []network.ConfigOption{
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	}
	opts = append(opts, options...)
	integrationNetwork := network.New(create, opts...)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)
	p256Priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).To(BeNil())

	suite := &IntegrationTestSuite{
		network:           integrationNetwork,
		factory:           txFactory,
		keyring:           keyring,
		precompileAddress: p256.Precompile{}.Address(),
		p256Priv:          p256Priv,
	}

	return suite
}
