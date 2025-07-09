package evidence

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SubmitEvidence implements the evidence submission logic for the evidence precompile.
func (p Precompile) SubmitEvidence(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, submitterHexAddr, err := NewMsgSubmitEvidence(method, args)
	if err != nil {
		return nil, err
	}
	router, err := p.evidenceKeeper.GetEvidenceHandler(msg.GetEvidence().Route())
	if err != nil {
		return nil, fmt.Errorf("failed to get evidence handler: %w", err)
	}
	if router == nil {
		return nil, fmt.Errorf("no evidence handler found for route: %s", msg.GetEvidence().Route())
	}

	msgSender := contract.Caller()
	if msgSender != submitterHexAddr {
		return nil, fmt.Errorf(cmn.ErrRequesterIsNotMsgSender, msgSender.String(), submitterHexAddr.String())
	}

	msgServer := evidencekeeper.NewMsgServerImpl(p.evidenceKeeper)
	res, err := msgServer.SubmitEvidence(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitSubmitEvidenceEvent(ctx, stateDB, submitterHexAddr, res.Hash); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
