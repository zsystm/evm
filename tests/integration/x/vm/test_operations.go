package vm

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	"github.com/cosmos/evm/x/vm/keeper/testdata"
	"github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
)

func (s *KeeperTestSuite) TestERC20() {
	s.MintFeeCollector = true
	s.SetupTest()

	// Fund fee collector with sufficient balance for gas refunds
	ctx := s.Network.GetContext()

	erc20Contract, err := testdata.LoadERC20Contract()
	s.Require().NoError(err)
	deployer := s.Keyring.GetAddr(0)

	// Deploy the ERC20 contract
	constructorArgs := []interface{}{
		deployer,
		sdkmath.NewIntWithDecimal(10000, 18).BigInt(),
	}

	contractAddr, err := s.Factory.DeployContract(
		s.Keyring.GetPrivKey(0),
		types.EvmTxArgs{
			ChainID:  s.Network.GetEIP155ChainID(),
			GasPrice: big.NewInt(1000000000),
		}, // Default values
		testutiltypes.ContractDeploymentData{
			Contract:        erc20Contract,
			ConstructorArgs: constructorArgs,
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network.NextBlock())

	// Generate random recipient address
	randomRecipient := common.BytesToAddress([]byte("randomRecipient123"))

	// Check initial balances
	k := s.Network.App.GetEVMKeeper()

	// Transfer tokens to random recipient
	transferAmount := big.NewInt(1000)
	transferInput, err := erc20Contract.ABI.Pack("transfer", randomRecipient, transferAmount)
	s.Require().NoError(err)

	nonce := k.GetNonce(ctx, deployer)
	transferArgs := types.EvmTxArgs{
		ChainID:  s.Network.GetEIP155ChainID(),
		Nonce:    nonce,
		To:       &contractAddr,
		Amount:   big.NewInt(0),
		GasLimit: 100000,
		GasPrice: big.NewInt(1000000000),
		Input:    transferInput,
	}

	fmt.Println("ERC20 Transfer")
	execResult, err := s.Factory.ExecuteEthTx(s.Keyring.GetPrivKey(0), transferArgs)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), execResult.Code, "Expected successful transfer execution")
}
