package example_chain

import (
	"github.com/cosmos/evm/example_chain/eips"
	"github.com/cosmos/evm/x/vm/core/vm"
)

// cosmosEVMActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var cosmosEVMActivators = map[string]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}
