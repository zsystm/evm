package types

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"

	sdkmath "cosmossdk.io/math"
)

// TransactionArgs represents the arguments to construct a new transaction
// or a message call using JSON-RPC.
// Duplicate struct definition since geth struct is in internal package
// Ref: https://github.com/ethereum/go-ethereum/blob/release/1.10.4/internal/ethapi/transaction_args.go#L36
type TransactionArgs struct {
	From                 *common.Address `json:"from"`
	To                   *common.Address `json:"to"`
	Gas                  *hexutil.Uint64 `json:"gas"`
	GasPrice             *hexutil.Big    `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas"`
	Value                *hexutil.Big    `json:"value"`
	Nonce                *hexutil.Uint64 `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://github.com/ethereum/go-ethereum/issues/15628
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`

	// Introduced by AccessListTxType transaction.
	AccessList *ethtypes.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big         `json:"chainId,omitempty"`

	// For BlobTxType
	BlobFeeCap *hexutil.Big  `json:"maxFeePerBlobGas"`
	BlobHashes []common.Hash `json:"blobVersionedHashes,omitempty"`

	// For BlobTxType transactions with blob sidecar
	Blobs       []kzg4844.Blob       `json:"blobs"`
	Commitments []kzg4844.Commitment `json:"commitments"`
	Proofs      []kzg4844.Proof      `json:"proofs"`

	// For SetCodeTxType
	AuthorizationList []ethtypes.SetCodeAuthorization `json:"authorizationList"`
}

// String return the struct in a string format
func (args *TransactionArgs) String() string {
	// Todo: There is currently a bug with hexutil.Big when the value its nil, printing would trigger an exception
	return fmt.Sprintf("TransactionArgs{From:%v, To:%v, Gas:%v,"+
		" Nonce:%v, Data:%v, Input:%v, AccessList:%v}",
		args.From,
		args.To,
		args.Gas,
		args.Nonce,
		args.Data,
		args.Input,
		args.AccessList)
}

// ToTransaction converts the arguments to an ethereum transaction.
// This assumes that setTxDefaults has been called.
func (args *TransactionArgs) ToTransaction() *MsgEthereumTx {
	var (
		chainID, value, gasPrice, maxFeePerGas, maxPriorityFeePerGas sdkmath.Int
		gas, nonce                                                   uint64
		from                                                         []byte
		to                                                           string
	)

	// Set sender address or use zero address if none specified.
	if args.ChainID != nil {
		chainID = sdkmath.NewIntFromBigInt(args.ChainID.ToInt())
	}

	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	}

	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}

	if args.GasPrice != nil {
		gasPrice = sdkmath.NewIntFromBigInt(args.GasPrice.ToInt())
	}

	if args.MaxFeePerGas != nil {
		maxFeePerGas = sdkmath.NewIntFromBigInt(args.MaxFeePerGas.ToInt())
	}

	if args.MaxPriorityFeePerGas != nil {
		maxPriorityFeePerGas = sdkmath.NewIntFromBigInt(args.MaxPriorityFeePerGas.ToInt())
	}

	if args.Value != nil {
		value = sdkmath.NewIntFromBigInt(args.Value.ToInt())
	}

	if args.To != nil {
		to = args.To.Hex()
	}

	var data TxData
	switch {
	case args.MaxFeePerGas != nil:
		al := AccessList{}
		if args.AccessList != nil {
			al = NewAccessList(args.AccessList)
		}

		data = &DynamicFeeTx{
			To:        to,
			ChainID:   &chainID,
			Nonce:     nonce,
			GasLimit:  gas,
			GasFeeCap: &maxFeePerGas,
			GasTipCap: &maxPriorityFeePerGas,
			Amount:    &value,
			Data:      args.GetData(),
			Accesses:  al,
		}
	case args.AccessList != nil:
		data = &AccessListTx{
			To:       to,
			ChainID:  &chainID,
			Nonce:    nonce,
			GasLimit: gas,
			GasPrice: &gasPrice,
			Amount:   &value,
			Data:     args.GetData(),
			Accesses: NewAccessList(args.AccessList),
		}
	default:
		data = &LegacyTx{
			To:       to,
			Nonce:    nonce,
			GasLimit: gas,
			GasPrice: &gasPrice,
			Amount:   &value,
			Data:     args.GetData(),
		}
	}

	anyData, err := PackTxData(data)
	if err != nil {
		return nil
	}

	if args.From != nil {
		from = args.From.Bytes()
	}

	msg := MsgEthereumTx{
		Data: anyData,
		From: from,
	}
	msg.Hash = msg.AsTransaction().Hash().Hex()
	return &msg
}

// ToMessage converts the arguments to the Message type used by the core evm.
// This assumes that setTxDefaults has been called.
func (args *TransactionArgs) ToMessage(globalGasCap uint64, baseFee *big.Int, skipNonceCheck,
	skipEoACheck bool,
) (core.Message, error) {
	// Reject invalid combinations of pre- and post-1559 fee styles
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return core.Message{}, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}

	// Set sender address or use zero address if none specified.
	addr := args.GetFrom()

	// Set default gas & gas price if none were set
	gas := globalGasCap
	if gas == 0 {
		gas = uint64(math.MaxUint64 / 2)
	}
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != 0 && globalGasCap < gas {
		gas = globalGasCap
	}
	var (
		gasPrice  *big.Int
		gasFeeCap *big.Int
		gasTipCap *big.Int
	)
	if baseFee == nil {
		// If there's no basefee, then it must be a non-1559 execution
		gasPrice = new(big.Int)
		if args.GasPrice != nil {
			gasPrice = args.GasPrice.ToInt()
		}
		gasFeeCap, gasTipCap = gasPrice, gasPrice
	} else {
		// A basefee is provided, necessitating 1559-type execution
		if args.GasPrice != nil {
			// User specified the legacy gas field, convert to 1559 gas typing
			gasPrice = args.GasPrice.ToInt()
			gasFeeCap, gasTipCap = gasPrice, gasPrice
		} else {
			// User specified 1559 gas feilds (or none), use those
			gasFeeCap = new(big.Int)
			if args.MaxFeePerGas != nil {
				gasFeeCap = args.MaxFeePerGas.ToInt()
			}
			gasTipCap = new(big.Int)
			if args.MaxPriorityFeePerGas != nil {
				gasTipCap = args.MaxPriorityFeePerGas.ToInt()
			}
			// Backfill the legacy gasPrice for EVM execution, unless we're all zeroes
			gasPrice = new(big.Int)
			if gasFeeCap.BitLen() > 0 || gasTipCap.BitLen() > 0 {
				gasPrice = new(big.Int).Add(gasTipCap, baseFee)
				if gasPrice.Cmp(gasFeeCap) > 0 {
					gasPrice = gasFeeCap
				}
			}
		}
	}
	value := new(big.Int)
	if args.Value != nil {
		value = args.Value.ToInt()
	}
	data := args.GetData()
	var accessList ethtypes.AccessList
	if args.AccessList != nil {
		accessList = *args.AccessList
	}

	nonce := uint64(0)
	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	}

	msg := core.Message{
		From:                  addr,
		To:                    args.To,
		Nonce:                 nonce,
		Value:                 value,
		GasLimit:              gas,
		GasPrice:              gasPrice,
		GasFeeCap:             gasFeeCap,
		GasTipCap:             gasTipCap,
		Data:                  data,
		AccessList:            accessList,
		BlobGasFeeCap:         (*big.Int)(args.BlobFeeCap),
		BlobHashes:            args.BlobHashes,
		SetCodeAuthorizations: args.AuthorizationList,
		SkipNonceChecks:       skipNonceCheck,
		SkipFromEOACheck:      skipEoACheck,
	}
	return msg, nil
}

// GetFrom retrieves the transaction sender address.
func (args *TransactionArgs) GetFrom() common.Address {
	if args.From == nil {
		return common.Address{}
	}
	return *args.From
}

// GetData retrieves the transaction calldata. Input field is preferred.
func (args *TransactionArgs) GetData() []byte {
	if args.Input != nil {
		return *args.Input
	}
	if args.Data != nil {
		return *args.Data
	}
	return nil
}
