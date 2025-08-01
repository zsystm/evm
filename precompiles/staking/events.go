package staking

import (
	"bytes"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// EventTypeCreateValidator defines the event type for the staking CreateValidator transaction.
	EventTypeCreateValidator = "CreateValidator"
	// EventTypeEditValidator defines the event type for the staking EditValidator transaction.
	EventTypeEditValidator = "EditValidator"
	// EventTypeDelegate defines the event type for the staking Delegate transaction.
	EventTypeDelegate = "Delegate"
	// EventTypeUnbond defines the event type for the staking Undelegate transaction.
	EventTypeUnbond = "Unbond"
	// EventTypeRedelegate defines the event type for the staking Redelegate transaction.
	EventTypeRedelegate = "Redelegate"
	// EventTypeCancelUnbondingDelegation defines the event type for the staking CancelUnbondingDelegation transaction.
	EventTypeCancelUnbondingDelegation = "CancelUnbondingDelegation"
)

// EmitCreateValidatorEvent creates a new create validator event emitted on a CreateValidator transaction.
func (p Precompile) EmitCreateValidatorEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgCreateValidator, validatorAddr common.Address) error {
	// Prepare the event topics
	event := p.Events[EventTypeCreateValidator]

	topics, err := p.createEditValidatorTxTopics(2, event, validatorAddr)
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(msg.Value.Amount.BigInt())))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// EmitEditValidatorEvent creates a new edit validator event emitted on a EditValidator transaction.
func (p Precompile) EmitEditValidatorEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgEditValidator, validatorAddr common.Address) error {
	// Prepare the event topics
	event := p.Events[EventTypeEditValidator]

	topics, err := p.createEditValidatorTxTopics(2, event, validatorAddr)
	if err != nil {
		return err
	}

	commissionRate := big.NewInt(DoNotModifyCommissionRate)
	if msg.CommissionRate != nil {
		commissionRate = msg.CommissionRate.BigInt()
	}

	minSelfDelegation := big.NewInt(DoNotModifyMinSelfDelegation)
	if msg.MinSelfDelegation != nil {
		minSelfDelegation = msg.MinSelfDelegation.BigInt()
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(commissionRate)))
	b.Write(cmn.PackNum(reflect.ValueOf(minSelfDelegation)))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// EmitDelegateEvent creates a new delegate event emitted on a Delegate transaction.
func (p Precompile) EmitDelegateEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgDelegate, delegatorAddr common.Address) error {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	// Get the validator to estimate the new shares delegated
	// NOTE: At this point the validator has already been checked, so no need to check again
	validator, _ := p.stakingKeeper.GetValidator(ctx, valAddr)

	// Get only the new shares based on the delegation amount
	newShares, err := validator.SharesFromTokens(msg.Amount.Amount)
	if err != nil {
		return err
	}

	// Prepare the event topics
	event := p.Events[EventTypeDelegate]
	topics, err := p.createStakingTxTopics(3, event, delegatorAddr, common.BytesToAddress(valAddr.Bytes()))
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(msg.Amount.Amount.BigInt())))
	b.Write(cmn.PackNum(reflect.ValueOf(newShares.BigInt())))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// EmitUnbondEvent creates a new unbond event emitted on an Undelegate transaction.
func (p Precompile) EmitUnbondEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgUndelegate, delegatorAddr common.Address, completionTime int64) error {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	// Prepare the event topics
	event := p.Events[EventTypeUnbond]
	topics, err := p.createStakingTxTopics(3, event, delegatorAddr, common.BytesToAddress(valAddr.Bytes()))
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(msg.Amount.Amount.BigInt())))
	b.Write(cmn.PackNum(reflect.ValueOf(big.NewInt(completionTime))))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// EmitRedelegateEvent creates a new redelegate event emitted on a Redelegate transaction.
func (p Precompile) EmitRedelegateEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgBeginRedelegate, delegatorAddr common.Address, completionTime int64) error {
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return err
	}

	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return err
	}

	// Prepare the event topics
	event := p.Events[EventTypeRedelegate]
	topics, err := p.createStakingTxTopics(4, event, delegatorAddr, common.BytesToAddress(valSrcAddr.Bytes()))
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(common.BytesToAddress(valDstAddr.Bytes()))
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(msg.Amount.Amount.BigInt())))
	b.Write(cmn.PackNum(reflect.ValueOf(big.NewInt(completionTime))))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// EmitCancelUnbondingDelegationEvent creates a new cancel unbonding delegation event emitted on a CancelUnbondingDelegation transaction.
func (p Precompile) EmitCancelUnbondingDelegationEvent(ctx sdk.Context, stateDB vm.StateDB, msg *stakingtypes.MsgCancelUnbondingDelegation, delegatorAddr common.Address) error {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	// Prepare the event topics
	event := p.Events[EventTypeCancelUnbondingDelegation]
	topics, err := p.createStakingTxTopics(3, event, delegatorAddr, common.BytesToAddress(valAddr.Bytes()))
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(msg.Amount.Amount.BigInt())))
	b.Write(cmn.PackNum(reflect.ValueOf(big.NewInt(msg.CreationHeight))))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115 // won't exceed uint64
	})

	return nil
}

// createStakingTxTopics creates the topics for staking transactions Delegate, Undelegate, Redelegate and CancelUnbondingDelegation.
func (p Precompile) createStakingTxTopics(topicsLen uint64, event abi.Event, delegatorAddr common.Address, validatorAddr common.Address) ([]common.Hash, error) {
	topics := make([]common.Hash, topicsLen)
	// NOTE: If your solidity event contains indexed event types, then they become a topic rather than part of the data property of the log.
	// In solidity you may only have up to 4 topics but only 3 indexed event types. The first topic is always the signature of the event.

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(delegatorAddr)
	if err != nil {
		return nil, err
	}

	topics[2], err = cmn.MakeTopic(validatorAddr)
	if err != nil {
		return nil, err
	}

	return topics, nil
}

// createEditValidatorTxTopics creates the topics for staking transactions CreateValidator and EditValidator.
func (p Precompile) createEditValidatorTxTopics(topicsLen uint64, event abi.Event, validatorAddr common.Address) ([]common.Hash, error) {
	topics := make([]common.Hash, topicsLen)
	// NOTE: If your solidity event contains indexed event types, then they become a topic rather than part of the data property of the log.
	// In solidity you may only have up to 4 topics but only 3 indexed event types. The first topic is always the signature of the event.

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(validatorAddr)
	if err != nil {
		return nil, err
	}

	return topics, nil
}
