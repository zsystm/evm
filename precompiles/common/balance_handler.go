package common

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"

	"github.com/cosmos/evm/utils"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BalanceHandler wraps bank accesses and records differences automatically.
type BalanceHandler struct {
	prevEventsLen int
}

func NewBalanceHandler() *BalanceHandler {
	return &BalanceHandler{
		prevEventsLen: 0,
	}
}

// Begin records initial balances for the provided addresses.
func (bh *BalanceHandler) BeforeBalanceChange(ctx sdk.Context) {
	bh.prevEventsLen = len(ctx.EventManager().Events())
}

// End compares balances and populates the journal.
func (bh *BalanceHandler) AfterBalanceChange(ctx sdk.Context, stateDB *statedb.StateDB) error {
	events := ctx.EventManager().Events()

	for _, event := range events[bh.prevEventsLen:] {
		switch event.Type {
		case banktypes.EventTypeCoinSpent:
			spenderHexAddr, err := parseHexAddress(event, banktypes.AttributeKeySpender)
			if err != nil {
				return fmt.Errorf("failed to parse spender address from event %q: %w", banktypes.EventTypeCoinSpent, err)
			}

			amount, err := parseAmount(event)
			if err != nil {
				return fmt.Errorf("failed to parse amount from event %q: %w", banktypes.EventTypeCoinSpent, err)
			}

			stateDB.SubBalance(spenderHexAddr, amount, tracing.BalanceChangeUnspecified)

		case banktypes.EventTypeCoinReceived:
			receiverHexAddr, err := parseHexAddress(event, banktypes.AttributeKeyReceiver)
			if err != nil {
				return fmt.Errorf("failed to parse receiver address from event %q: %w", banktypes.EventTypeCoinReceived, err)
			}

			amount, err := parseAmount(event)
			if err != nil {
				return fmt.Errorf("failed to parse amount from event %q: %w", banktypes.EventTypeCoinReceived, err)
			}

			stateDB.AddBalance(receiverHexAddr, amount, tracing.BalanceChangeUnspecified)
		}
	}

	return nil
}

func parseHexAddress(event sdk.Event, key string) (common.Address, error) {
	attr, ok := event.GetAttribute(key)
	if !ok {
		return common.Address{}, fmt.Errorf("event %q missing attribute %q", event.Type, key)
	}

	accAddr, err := sdk.AccAddressFromBech32(attr.Value)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid address %q: %w", attr.Value, err)
	}

	return common.Address(accAddr.Bytes()), nil
}

func parseAmount(event sdk.Event) (*uint256.Int, error) {
	amountAttr, ok := event.GetAttribute(sdk.AttributeKeyAmount)
	if !ok {
		return nil, fmt.Errorf("event %q missing attribute %q", banktypes.EventTypeCoinSpent, sdk.AttributeKeyAmount)
	}

	amountCoins, err := sdk.ParseCoinsNormalized(amountAttr.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coins from %q: %w", amountAttr.Value, err)
	}

	amountBigInt := amountCoins.AmountOf(evmtypes.GetEVMCoinDenom()).BigInt()
	amount, err := utils.Uint256FromBigInt(evmtypes.ConvertAmountTo18DecimalsBigInt(amountBigInt))
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin amount to Uint256: %w", err)
	}
	return amount, nil
}
