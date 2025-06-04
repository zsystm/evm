package evm

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	precisebankkeeper "github.com/cosmos/evm/x/precisebank/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// EvmApp defines the interface for an EVM application.
type EvmApp interface {
	ibctesting.TestingApp
	runtime.AppI
	InterfaceRegistry() types.InterfaceRegistry
	ChainID() string
	GetEVMKeeper() *evmkeeper.Keeper
	GetErc20Keeper() Erc20Keeper
	SetErc20Keeper(Erc20Keeper)
	GetGovKeeper() govkeeper.Keeper
	GetSlashingKeeper() slashingkeeper.Keeper
	GetEvidenceKeeper() *evidencekeeper.Keeper
	GetBankKeeper() bankkeeper.Keeper
	GetFeeMarketKeeper() *feemarketkeeper.Keeper
	GetAccountKeeper() authkeeper.AccountKeeper
	GetAuthzKeeper() authzkeeper.Keeper
	GetDistrKeeper() distrkeeper.Keeper
	GetStakingKeeper() *stakingkeeper.Keeper
	GetMintKeeper() mintkeeper.Keeper
	GetPreciseBankKeeper() *precisebankkeeper.Keeper
	GetFeeGrantKeeper() feegrantkeeper.Keeper
	GetTransferKeeper() transferkeeper.Keeper
	SetTransferKeeper(transferKeeper transferkeeper.Keeper)
	DefaultGenesis() map[string]json.RawMessage
	GetKey(storeKey string) *storetypes.KVStoreKey
	GetAnteHandler() sdk.AnteHandler
	GetSubspace(moduleName string) paramstypes.Subspace
	MsgServiceRouter() *baseapp.MsgServiceRouter
}

type Erc20Keeper interface {
	SetToken(ctx sdk.Context, pair erc20types.TokenPair)
	GetTokenPairID(ctx sdk.Context, token string) []byte
	GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool)
	GetParams(ctx sdk.Context) erc20types.Params
	SetAllowance(ctx sdk.Context, erc20 common.Address, owner common.Address, spender common.Address, value *big.Int) error
	GetAllowance(ctx sdk.Context, erc20 common.Address, owner common.Address, spender common.Address) (*big.Int, error)
	DeleteAllowance(ctx sdk.Context, erc20 common.Address, owner common.Address, spender common.Address) error
	EnableDynamicPrecompiles(ctx sdk.Context, addresses ...common.Address) error
	TokenPairs(context.Context, *erc20types.QueryTokenPairsRequest) (*erc20types.QueryTokenPairsResponse, error)
	TokenPair(context.Context, *erc20types.QueryTokenPairRequest) (*erc20types.QueryTokenPairResponse, error)
	Params(context.Context, *erc20types.QueryParamsRequest) (*erc20types.QueryParamsResponse, error)
	RegisterERC20Extension(ctx sdk.Context, denom string) (*erc20types.TokenPair, error)
	GetDenomMap(ctx sdk.Context, denom string) []byte
	GetERC20Map(ctx sdk.Context, erc20 common.Address) []byte
	BalanceOf(ctx sdk.Context, abi abi.ABI, contract, account common.Address) *big.Int
	GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error)
	GetTokenDenom(ctx sdk.Context, tokenAddress common.Address) (string, error)
	ConvertERC20(ctx context.Context, msg *erc20types.MsgConvertERC20) (*erc20types.MsgConvertERC20Response, error)
	IsERC20Enabled(ctx sdk.Context) bool
	GetTokenPairs(ctx sdk.Context) []erc20types.TokenPair
	GetAllowances(ctx sdk.Context) []erc20types.Allowance
	SetERC20Map(ctx sdk.Context, erc20 common.Address, id []byte)
	SetDenomMap(ctx sdk.Context, denom string, id []byte)
	CreateCoinMetadata(ctx sdk.Context, contract common.Address) (*banktypes.Metadata, error)
	SetPermissionlessRegistration(ctx sdk.Context, permissionlessRegistration bool)
	RegisterERC20(goCtx context.Context, req *erc20types.MsgRegisterERC20) (*erc20types.MsgRegisterERC20Response, error)
	ToggleConversion(goCtx context.Context, req *erc20types.MsgToggleConversion) (*erc20types.MsgToggleConversionResponse, error)
	UnsafeSetAllowance(
		ctx sdk.Context,
		erc20 common.Address,
		owner common.Address,
		spender common.Address,
		value *big.Int,
	) error
	DeleteTokenPair(ctx sdk.Context, tokenPair erc20types.TokenPair)
	RegisterERC20CodeHash(ctx sdk.Context, erc20Addr common.Address) error
	UnRegisterERC20CodeHash(ctx sdk.Context, erc20Addr common.Address) error
	QueryERC20(
		ctx sdk.Context,
		contract common.Address,
	) (erc20types.ERC20Data, error)
	SetTokenPair(ctx sdk.Context, tokenPair erc20types.TokenPair)
	SetParams(ctx sdk.Context, newParams erc20types.Params) error
	ConvertCoinToERC20FromPacket(ctx sdk.Context, data transfertypes.FungibleTokenPacketData) error
	OnRecvPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		ack exported.Acknowledgement,
	) exported.Acknowledgement
	OnAcknowledgementPacket(
		ctx sdk.Context, _ channeltypes.Packet,
		data transfertypes.FungibleTokenPacketData,
		ack channeltypes.Acknowledgement,
	) error
	OnTimeoutPacket(ctx sdk.Context, _ channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error
	MintingEnabled(
		ctx sdk.Context,
		sender, receiver sdk.AccAddress,
		token string,
	) (erc20types.TokenPair, error)
	ConvertCoin(
		goCtx context.Context,
		msg *erc20types.MsgConvertCoin,
	) (*erc20types.MsgConvertCoinResponse, error)
	UpdateParams(goCtx context.Context, req *erc20types.MsgUpdateParams) (*erc20types.MsgUpdateParamsResponse, error)
	GetERC20PrecompileInstance(
		ctx sdk.Context,
		address common.Address,
	) (contract vm.PrecompiledContract, found bool, err error)
	IsTokenPairRegistered(ctx sdk.Context, id []byte) bool
	IsERC20Registered(ctx sdk.Context, erc20 common.Address) bool
	IsDenomRegistered(ctx sdk.Context, denom string) bool
}
