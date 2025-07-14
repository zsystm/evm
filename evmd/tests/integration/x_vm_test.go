package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/x/vm"
)

func TestKeeperTestSuite(t *testing.T) {
	s := vm.NewKeeperTestSuite(CreateEvmd)
	s.EnableFeemarket = false
	s.EnableLondonHF = true
	suite.Run(t, s)
}

func TestNestedEVMExtensionCallSuite(t *testing.T) {
	s := vm.NewNestedEVMExtensionCallSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestGenesisTestSuite(t *testing.T) {
	s := vm.NewGenesisTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestVmAnteTestSuite(t *testing.T) {
	s := vm.NewEvmAnteTestSuite(CreateEvmd)
	suite.Run(t, s)
}

func TestIterateContracts(t *testing.T) {
	vm.TestIterateContracts(t, CreateEvmd)
}

func BenchmarkApplyTransaction(b *testing.B) {
	vm.BenchmarkApplyTransaction(b, CreateEvmd)
}

func BenchmarkApplyTransactionWithLegacyTx(b *testing.B) {
	vm.BenchmarkApplyTransactionWithLegacyTx(b, CreateEvmd)
}

func BenchmarkApplyTransactionWithDynamicFeeTx(b *testing.B) {
	vm.BenchmarkApplyTransactionWithDynamicFeeTx(b, CreateEvmd)
}

func BenchmarkApplyMessage(b *testing.B) {
	vm.BenchmarkApplyMessage(b, CreateEvmd)
}

func BenchmarkApplyMessageWithLegacyTx(b *testing.B) {
	vm.BenchmarkApplyMessageWithLegacyTx(b, CreateEvmd)
}

func BenchmarkApplyMessageWithDynamicFeeTx(b *testing.B) {
	vm.BenchmarkApplyMessageWithDynamicFeeTx(b, CreateEvmd)
}
