// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package ethereum implements execution of Ethereum JSON tests for cosmos/evm.
package ethereum

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/evm/testutil/constants"
	stdmath "math"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/ethereum/go-ethereum/triedb/pathdb"

	// Cosmos/EVM imports for actual state transitions
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	cosmosmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	evmd "github.com/cosmos/evm/evmd"
	"github.com/cosmos/evm/testutil/integration/evm/network"
)

// A BlockTest checks handling of entire blocks.
type BlockTest struct {
	json btJSON
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (t *BlockTest) UnmarshalJSON(in []byte) error {
	return json.Unmarshal(in, &t.json)
}

type btJSON struct {
	Blocks     []btBlock             `json:"blocks"`
	Genesis    btHeader              `json:"genesisBlockHeader"`
	Pre        types.GenesisAlloc    `json:"pre"`
	Post       types.GenesisAlloc    `json:"postState"`
	BestBlock  common.UnprefixedHash `json:"lastblockhash"`
	Network    string                `json:"network"`
	SealEngine string                `json:"sealEngine"`
}

type btBlock struct {
	BlockHeader     *btHeader
	ExpectException string
	Rlp             string
	UncleHeaders    []*btHeader
}

//go:generate go run github.com/fjl/gencodec -type btHeader -field-override btHeaderMarshaling -out gen_btheader.go

type btHeader struct {
	Bloom                 types.Bloom
	Coinbase              common.Address
	MixHash               common.Hash
	Nonce                 types.BlockNonce
	Number                *big.Int
	Hash                  common.Hash
	ParentHash            common.Hash
	ReceiptTrie           common.Hash
	StateRoot             common.Hash
	TransactionsTrie      common.Hash
	UncleHash             common.Hash
	ExtraData             []byte
	Difficulty            *big.Int
	GasLimit              uint64
	GasUsed               uint64
	Timestamp             uint64
	BaseFeePerGas         *big.Int
	WithdrawalsRoot       *common.Hash
	BlobGasUsed           *uint64
	ExcessBlobGas         *uint64
	ParentBeaconBlockRoot *common.Hash
}

type btHeaderMarshaling struct {
	ExtraData     hexutil.Bytes
	Number        *math.HexOrDecimal256
	Difficulty    *math.HexOrDecimal256
	GasLimit      math.HexOrDecimal64
	GasUsed       math.HexOrDecimal64
	Timestamp     math.HexOrDecimal64
	BaseFeePerGas *math.HexOrDecimal256
	BlobGasUsed   *math.HexOrDecimal64
	ExcessBlobGas *math.HexOrDecimal64
}

// BlockTestSuite provides the setup for Ethereum Block Test execution using cosmos/evm
type BlockTestSuite struct {
	network network.Network
	create  network.CreateEvmApp
	options []network.ConfigOption
}

// NewBlockTestSuite creates a new BlockTestSuite for running Ethereum block tests
func NewBlockTestSuite(create network.CreateEvmApp, options ...network.ConfigOption) *BlockTestSuite {
	return &BlockTestSuite{
		create:  create,
		options: options,
	}
}

// SetupTest initializes the test network for block test execution
func (s *BlockTestSuite) SetupTest() {
	// Use unit test network configuration
	nw := network.NewUnitTestNetwork(s.create, s.options...)
	s.network = nw
}

func (t *BlockTest) Run(snapshotter bool, scheme string, witness bool, tracer *tracing.Hooks, postCheck func(error, *core.BlockChain)) (result error) {
	config, ok := Forks[t.json.Network]
	if !ok {
		return UnsupportedForkError{t.json.Network}
	}
	// import pre accounts & construct test genesis block & state root
	var (
		db    = rawdb.NewMemoryDatabase()
		tconf = &triedb.Config{
			Preimages: true,
		}
	)
	if scheme == rawdb.PathScheme {
		tconf.PathDB = pathdb.Defaults
	} else {
		tconf.HashDB = hashdb.Defaults
	}
	// Commit genesis state
	gspec := t.genesis(config)

	// if ttd is not specified, set an arbitrary huge value
	if gspec.Config.TerminalTotalDifficulty == nil {
		gspec.Config.TerminalTotalDifficulty = big.NewInt(stdmath.MaxInt64)
	}
	triedb := triedb.NewDatabase(db, tconf)
	gblock, err := gspec.Commit(db, triedb)
	if err != nil {
		return err
	}
	triedb.Close() // close the db to prevent memory leak

	if gblock.Hash() != t.json.Genesis.Hash {
		return fmt.Errorf("genesis block hash doesn't match test: computed=%x, test=%x", gblock.Hash().Bytes()[:6], t.json.Genesis.Hash[:6])
	}
	if gblock.Root() != t.json.Genesis.StateRoot {
		return fmt.Errorf("genesis block state root does not match test: computed=%x, test=%x", gblock.Root().Bytes()[:6], t.json.Genesis.StateRoot[:6])
	}
	// Wrap the original engine within the beacon-engine
	engine := beacon.New(ethash.NewFaker())

	cache := &core.CacheConfig{TrieCleanLimit: 0, StateScheme: scheme, Preimages: true}
	if snapshotter {
		cache.SnapshotLimit = 1
		cache.SnapshotWait = true
	}
	chain, err := core.NewBlockChain(db, cache, gspec, nil, engine, vm.Config{
		Tracer:                  tracer,
		StatelessSelfValidation: witness,
	}, nil)
	if err != nil {
		return err
	}
	defer chain.Stop()

	validBlocks, err := t.insertBlocks(chain)
	if err != nil {
		return err
	}
	// Import succeeded: regardless of whether the _test_ succeeds or not, schedule
	// the post-check to run
	if postCheck != nil {
		defer postCheck(result, chain)
	}
	cmlast := chain.CurrentBlock().Hash()
	if common.Hash(t.json.BestBlock) != cmlast {
		return fmt.Errorf("last block hash validation mismatch: want: %x, have: %x", t.json.BestBlock, cmlast)
	}
	newDB, err := chain.State()
	if err != nil {
		return err
	}
	if err = t.validatePostState(newDB); err != nil {
		return fmt.Errorf("post state validation failed: %v", err)
	}
	// Cross-check the snapshot-to-hash against the trie hash
	if snapshotter {
		if err := chain.Snapshots().Verify(chain.CurrentBlock().Root); err != nil {
			return err
		}
	}
	return t.validateImportedHeaders(chain, validBlocks)
}

func (t *BlockTest) genesis(config *params.ChainConfig) *core.Genesis {
	return &core.Genesis{
		Config:        config,
		Nonce:         t.json.Genesis.Nonce.Uint64(),
		Timestamp:     t.json.Genesis.Timestamp,
		ParentHash:    t.json.Genesis.ParentHash,
		ExtraData:     t.json.Genesis.ExtraData,
		GasLimit:      t.json.Genesis.GasLimit,
		GasUsed:       t.json.Genesis.GasUsed,
		Difficulty:    t.json.Genesis.Difficulty,
		Mixhash:       t.json.Genesis.MixHash,
		Coinbase:      t.json.Genesis.Coinbase,
		Alloc:         t.json.Pre,
		BaseFee:       t.json.Genesis.BaseFeePerGas,
		BlobGasUsed:   t.json.Genesis.BlobGasUsed,
		ExcessBlobGas: t.json.Genesis.ExcessBlobGas,
	}
}

/*
See https://github.com/ethereum/tests/wiki/Blockchain-Tests-II

	Whether a block is valid or not is a bit subtle, it's defined by presence of
	blockHeader, transactions and uncleHeaders fields. If they are missing, the block is
	invalid and we must verify that we do not accept it.

	Since some tests mix valid and invalid blocks we need to check this for every block.

	If a block is invalid it does not necessarily fail the test, if it's invalidness is
	expected we are expected to ignore it and continue processing and then validate the
	post state.
*/
func (t *BlockTest) insertBlocks(blockchain *core.BlockChain) ([]btBlock, error) {
	validBlocks := make([]btBlock, 0)
	// insert the test blocks, which will execute all transactions
	for bi, b := range t.json.Blocks {
		cb, err := b.decode()
		if err != nil {
			if b.BlockHeader == nil {
				log.Info("Block decoding failed", "index", bi, "err", err)
				continue // OK - block is supposed to be invalid, continue with next block
			} else {
				return nil, fmt.Errorf("block RLP decoding failed when expected to succeed: %v", err)
			}
		}
		// RLP decoding worked, try to insert into chain:
		blocks := types.Blocks{cb}
		i, err := blockchain.InsertChain(blocks)
		if err != nil {
			if b.BlockHeader == nil {
				continue // OK - block is supposed to be invalid, continue with next block
			} else {
				return nil, fmt.Errorf("block #%v insertion into chain failed: %v", blocks[i].Number(), err)
			}
		}
		if b.BlockHeader == nil {
			if data, err := json.MarshalIndent(cb.Header(), "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "block (index %d) insertion should have failed due to: %v:\n%v\n",
					bi, b.ExpectException, string(data))
			}
			return nil, fmt.Errorf("block (index %d) insertion should have failed due to: %v",
				bi, b.ExpectException)
		}

		// validate RLP decoding by checking all values against test file JSON
		if err = validateHeader(b.BlockHeader, cb.Header()); err != nil {
			return nil, fmt.Errorf("deserialised block header validation failed: %v", err)
		}
		validBlocks = append(validBlocks, b)
	}
	return validBlocks, nil
}

func validateHeader(h *btHeader, h2 *types.Header) error {
	if h.Bloom != h2.Bloom {
		return fmt.Errorf("bloom: want: %x have: %x", h.Bloom, h2.Bloom)
	}
	if h.Coinbase != h2.Coinbase {
		return fmt.Errorf("coinbase: want: %x have: %x", h.Coinbase, h2.Coinbase)
	}
	if h.MixHash != h2.MixDigest {
		return fmt.Errorf("MixHash: want: %x have: %x", h.MixHash, h2.MixDigest)
	}
	if h.Nonce != h2.Nonce {
		return fmt.Errorf("nonce: want: %x have: %x", h.Nonce, h2.Nonce)
	}
	if h.Number.Cmp(h2.Number) != 0 {
		return fmt.Errorf("number: want: %v have: %v", h.Number, h2.Number)
	}
	if h.ParentHash != h2.ParentHash {
		return fmt.Errorf("parent hash: want: %x have: %x", h.ParentHash, h2.ParentHash)
	}
	if h.ReceiptTrie != h2.ReceiptHash {
		return fmt.Errorf("receipt hash: want: %x have: %x", h.ReceiptTrie, h2.ReceiptHash)
	}
	if h.TransactionsTrie != h2.TxHash {
		return fmt.Errorf("tx hash: want: %x have: %x", h.TransactionsTrie, h2.TxHash)
	}
	if h.StateRoot != h2.Root {
		return fmt.Errorf("state hash: want: %x have: %x", h.StateRoot, h2.Root)
	}
	if h.UncleHash != h2.UncleHash {
		return fmt.Errorf("uncle hash: want: %x have: %x", h.UncleHash, h2.UncleHash)
	}
	if !bytes.Equal(h.ExtraData, h2.Extra) {
		return fmt.Errorf("extra data: want: %x have: %x", h.ExtraData, h2.Extra)
	}
	if h.Difficulty.Cmp(h2.Difficulty) != 0 {
		return fmt.Errorf("difficulty: want: %v have: %v", h.Difficulty, h2.Difficulty)
	}
	if h.GasLimit != h2.GasLimit {
		return fmt.Errorf("gasLimit: want: %d have: %d", h.GasLimit, h2.GasLimit)
	}
	if h.GasUsed != h2.GasUsed {
		return fmt.Errorf("gasUsed: want: %d have: %d", h.GasUsed, h2.GasUsed)
	}
	if h.Timestamp != h2.Time {
		return fmt.Errorf("timestamp: want: %v have: %v", h.Timestamp, h2.Time)
	}
	if !reflect.DeepEqual(h.BaseFeePerGas, h2.BaseFee) {
		return fmt.Errorf("baseFeePerGas: want: %v have: %v", h.BaseFeePerGas, h2.BaseFee)
	}
	if !reflect.DeepEqual(h.WithdrawalsRoot, h2.WithdrawalsHash) {
		return fmt.Errorf("withdrawalsRoot: want: %v have: %v", h.WithdrawalsRoot, h2.WithdrawalsHash)
	}
	if !reflect.DeepEqual(h.BlobGasUsed, h2.BlobGasUsed) {
		return fmt.Errorf("blobGasUsed: want: %v have: %v", h.BlobGasUsed, h2.BlobGasUsed)
	}
	if !reflect.DeepEqual(h.ExcessBlobGas, h2.ExcessBlobGas) {
		return fmt.Errorf("excessBlobGas: want: %v have: %v", h.ExcessBlobGas, h2.ExcessBlobGas)
	}
	if !reflect.DeepEqual(h.ParentBeaconBlockRoot, h2.ParentBeaconRoot) {
		return fmt.Errorf("parentBeaconBlockRoot: want: %v have: %v", h.ParentBeaconBlockRoot, h2.ParentBeaconRoot)
	}
	return nil
}

func (t *BlockTest) validatePostState(statedb *state.StateDB) error {
	// validate post state accounts in test file against what we have in state db
	for addr, acct := range t.json.Post {
		// address is indirectly verified by the other fields, as it's the db key
		code2 := statedb.GetCode(addr)
		balance2 := statedb.GetBalance(addr).ToBig()
		nonce2 := statedb.GetNonce(addr)
		if !bytes.Equal(code2, acct.Code) {
			return fmt.Errorf("account code mismatch for addr: %s want: %v have: %s", addr, acct.Code, hex.EncodeToString(code2))
		}
		if balance2.Cmp(acct.Balance) != 0 {
			return fmt.Errorf("account balance mismatch for addr: %s, want: %d, have: %d", addr, acct.Balance, balance2)
		}
		if nonce2 != acct.Nonce {
			return fmt.Errorf("account nonce mismatch for addr: %s want: %d have: %d", addr, acct.Nonce, nonce2)
		}
		for k, v := range acct.Storage {
			v2 := statedb.GetState(addr, k)
			if v2 != v {
				return fmt.Errorf("account storage mismatch for addr: %s, slot: %x, want: %x, have: %x", addr, k, v, v2)
			}
		}
	}
	return nil
}

func (t *BlockTest) validateImportedHeaders(cm *core.BlockChain, validBlocks []btBlock) error {
	// to get constant lookup when verifying block headers by hash (some tests have many blocks)
	bmap := make(map[common.Hash]btBlock, len(t.json.Blocks))
	for _, b := range validBlocks {
		bmap[b.BlockHeader.Hash] = b
	}
	// iterate over blocks backwards from HEAD and validate imported
	// headers vs test file. some tests have reorgs, and we import
	// block-by-block, so we can only validate imported headers after
	// all blocks have been processed by BlockChain, as they may not
	// be part of the longest chain until last block is imported.
	for b := cm.CurrentBlock(); b != nil && b.Number.Uint64() != 0; b = cm.GetBlockByHash(b.ParentHash).Header() {
		if err := validateHeader(bmap[b.Hash()].BlockHeader, b); err != nil {
			return fmt.Errorf("imported block header validation failed: %v", err)
		}
	}
	return nil
}

func (bb *btBlock) decode() (*types.Block, error) {
	data, err := hexutil.Decode(bb.Rlp)
	if err != nil {
		return nil, err
	}
	var b types.Block
	err = rlp.DecodeBytes(data, &b)
	return &b, err
}

// RunWithEVMD executes the block test using cosmos/evm instead of geth
func (t *BlockTest) RunWithEVMD(test *testing.T, jsonData []byte) error {
	// Parse the JSON data into our test structure
	var btJSON btJSON
	if err := json.Unmarshal(jsonData, &btJSON); err != nil {
		return err
	}

	// Skip tests that cosmos/evm cannot handle
	if err := t.skipIncompatibleTest(btJSON); err != nil {
		test.Skipf("Skipping test incompatible with cosmos/evm: %v", err)
	}

	// Validate network configuration
	config, ok := Forks[btJSON.Network]
	if !ok {
		return UnsupportedForkError{btJSON.Network}
	}

	// Execute using cosmos/evm instead of geth
	app, valSet, err := t.setupEVMDApp(test, btJSON, config)
	if err != nil {
		return fmt.Errorf("failed to setup EVMD app: %v", err)
	}

	if err := t.executeBlocksWithEVMD(test, app, btJSON, valSet); err != nil {
		return fmt.Errorf("block execution failed: %v", err)
	}

	// Validate final state
	if err := t.validatePostStateWithEVMD(test, app, btJSON.Post); err != nil {
		return fmt.Errorf("post-state validation failed: %v", err)
	}

	return nil
}

// RunWithBlockTestSuite executes the block test using the BlockTestSuite and Network pattern
func (t *BlockTest) RunWithBlockTestSuite(suite *BlockTestSuite, test *testing.T, jsonData []byte) error {
	// Parse the JSON data into our test structure
	var btJSON btJSON
	if err := json.Unmarshal(jsonData, &btJSON); err != nil {
		return err
	}

	// Skip tests that cosmos/evm cannot handle
	if err := t.skipIncompatibleTest(btJSON); err != nil {
		test.Skipf("Skipping test incompatible with cosmos/evm: %v", err)
	}

	// Validate network configuration
	config, ok := Forks[btJSON.Network]
	if !ok {
		return UnsupportedForkError{btJSON.Network}
	}
	chainID := constants.ChainID{
		ChainID:    "cosmos-1",
		EVMChainID: config.ChainID.Uint64(),
	}

	// Create custom genesis state from the block test pre-state
	customGen := t.createCustomGenesisFromPreState(btJSON.Pre)

	// Create a new suite with custom genesis configuration
	blockSuite := NewBlockTestSuite(
		suite.create,
		append(
			suite.options,
			network.WithCustomGenesis(customGen),
			network.WithChainID(chainID),
		)...,
	)
	blockSuite.SetupTest()

	if err := t.executeBlocksWithSuite(blockSuite, btJSON); err != nil {
		return fmt.Errorf("block execution failed: %v", err)
	}

	// Validate final state
	if err := t.validatePostStateWithSuite(blockSuite, btJSON.Post); err != nil {
		return fmt.Errorf("post-state validation failed: %v", err)
	}

	return nil
}

// skipIncompatibleTest checks if a test should be skipped due to cosmos/evm incompatibility
func (t *BlockTest) skipIncompatibleTest(btJSON btJSON) error {
	// Skip tests with features that cosmos/evm doesn't support

	// Check for uncle/ommer blocks (PoW specific)
	for _, block := range btJSON.Blocks {
		if len(block.UncleHeaders) > 0 {
			return fmt.Errorf("test contains uncle/ommer blocks not supported by cosmos/evm")
		}
	}

	// Check for withdrawal operations (Beacon chain specific)
	if btJSON.Network == "Cancun" || btJSON.Network == "Shanghai" {
		for _, block := range btJSON.Blocks {
			if block.BlockHeader != nil && block.BlockHeader.WithdrawalsRoot != nil {
				return fmt.Errorf("test contains withdrawals not supported by cosmos/evm")
			}
		}
	}

	// Check for difficulty-based tests (PoW specific)
	if btJSON.Genesis.Difficulty != nil && btJSON.Genesis.Difficulty.Cmp(big.NewInt(0)) > 0 {
		return fmt.Errorf("test relies on PoW difficulty not supported by cosmos/evm")
	}

	return nil
}

// setupEVMDApp initializes a cosmos/evm application with the test genesis state
func (t *BlockTest) setupEVMDApp(test *testing.T, btJSON btJSON, config *params.ChainConfig) (*evmd.EVMD, *cmttypes.ValidatorSet, error) {
	test.Helper()

	// Create genesis accounts from the pre-state
	var genesisAccounts []authtypes.GenesisAccount
	var balances []banktypes.Balance

	for addr, account := range btJSON.Pre {
		// Convert Ethereum address to Cosmos address format
		cosmosAddr := sdk.AccAddress(addr.Bytes())

		// Create genesis account
		baseAccount := authtypes.NewBaseAccount(
			cosmosAddr,
			nil, // pubkey will be set when account sends first tx
			0,   // account number
			account.Nonce,
		)
		genesisAccounts = append(genesisAccounts, baseAccount)

		// Create balance entry
		if account.Balance != nil && account.Balance.Sign() > 0 {
			balance := banktypes.Balance{
				Address: cosmosAddr.String(),
				Coins: sdk.NewCoins(sdk.NewCoin(
					sdk.DefaultBondDenom, // Use standard denomination like test_helpers.go
					cosmosmath.NewIntFromBigInt(account.Balance),
				)),
			}
			balances = append(balances, balance)
		}
	}

	// Create validator set for the test (following test_helpers.go pattern)
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get validator pubkey: %v", err)
	}

	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// Initialize EVMD app with genesis state
	chainID := "test-chain"

	app := evmd.SetupWithGenesisValSet(test, chainID, config.ChainID.Uint64(), valSet, genesisAccounts, balances...)

	return app, valSet, nil
}

// executeBlocksWithEVMD executes the test blocks using cosmos/evm's EVMD
func (t *BlockTest) executeBlocksWithEVMD(test *testing.T, app *evmd.EVMD, btJSON btJSON, valSet *cmttypes.ValidatorSet) error {
	test.Helper()

	// Create initial context with proposer address
	ctx := app.NewContextLegacy(false, cmtproto.Header{
		ChainID:         "test-chain",
		Height:          1,
		Time:            time.Unix(int64(btJSON.Genesis.Timestamp), 0),
		ProposerAddress: valSet.Proposer.Address,
	})

	// Process each block
	for blockIndex, blockDef := range btJSON.Blocks {
		// Skip invalid blocks
		if blockDef.BlockHeader == nil {
			continue
		}

		// Decode the block to get transactions
		gethBlock, err := blockDef.decode()
		if err != nil {
			if blockDef.ExpectException != "" {
				continue
			}
			return fmt.Errorf("failed to decode block %d: %v", blockIndex, err)
		}

		// Execute transactions in the block
		for txIndex, tx := range gethBlock.Transactions() {
			// Convert Ethereum transaction to cosmos/evm message
			msgEthTx, err := t.convertToMsgEthereumTx(tx)
			if err != nil {
				return fmt.Errorf("failed to convert tx %d in block %d: %v", txIndex, blockIndex, err)
			}

			// Execute the transaction using EVMKeeper
			evmKeeper := app.GetEVMKeeper()
			_, err = evmKeeper.ApplyTransaction(ctx, msgEthTx)
			if err != nil {
				// Check if this failure was expected
				if blockDef.ExpectException != "" {
					continue
				}
				return fmt.Errorf("transaction execution failed: %v", err)
			}
		}
	}

	return nil
}

// convertToMsgEthereumTx converts a geth transaction to cosmos/evm MsgEthereumTx
func (t *BlockTest) convertToMsgEthereumTx(tx *types.Transaction) (*evmtypes.MsgEthereumTx, error) {
	// Create MsgEthereumTx from the Ethereum transaction
	msgEthTx := &evmtypes.MsgEthereumTx{}
	if err := msgEthTx.FromEthereumTx(tx); err != nil {
		return nil, fmt.Errorf("failed to convert ethereum tx: %v", err)
	}

	return msgEthTx, nil
}

// validatePostStateWithEVMD validates the final state against expected post-state
func (t *BlockTest) validatePostStateWithEVMD(test *testing.T, app *evmd.EVMD, expectedPostState types.GenesisAlloc) error {
	test.Helper()

	// Create context for state queries
	ctx := app.NewContextLegacy(false, cmtproto.Header{})
	evmKeeper := app.GetEVMKeeper()

	// Validate each expected account
	for addr, expectedAccount := range expectedPostState {

		// Get account balance
		cosmosAddr := sdk.AccAddress(addr.Bytes())
		balance := app.BankKeeper.GetBalance(ctx, cosmosAddr, sdk.DefaultBondDenom)
		actualBalance := balance.Amount.BigInt()

		if expectedAccount.Balance.Cmp(actualBalance) != 0 {
			return fmt.Errorf("balance mismatch for %s: expected %s, got %s",
				addr.Hex(), expectedAccount.Balance.String(), actualBalance.String())
		}

		// Get account nonce
		account := app.AccountKeeper.GetAccount(ctx, cosmosAddr)
		if account == nil {
			if expectedAccount.Nonce != 0 {
				return fmt.Errorf("account %s not found, but expected nonce %d", addr.Hex(), expectedAccount.Nonce)
			}
		} else {
			if account.GetSequence() != expectedAccount.Nonce {
				return fmt.Errorf("nonce mismatch for %s: expected %d, got %d",
					addr.Hex(), expectedAccount.Nonce, account.GetSequence())
			}
		}

		// Get account code (if any)
		codeHash := evmKeeper.GetCodeHash(ctx, addr)
		code := evmKeeper.GetCode(ctx, codeHash)
		if !bytes.Equal(code, expectedAccount.Code) {
			return fmt.Errorf("code mismatch for %s: expected %s, got %s",
				addr.Hex(), hex.EncodeToString(expectedAccount.Code), hex.EncodeToString(code))
		}

		// Validate storage (if any)
		for key, expectedValue := range expectedAccount.Storage {
			actualValue := evmKeeper.GetState(ctx, addr, key)
			if actualValue != expectedValue {
				return fmt.Errorf("storage mismatch for %s at key %s: expected %s, got %s",
					addr.Hex(), key.Hex(), expectedValue.Hex(), actualValue.Hex())
			}
		}
	}

	return nil
}

// createCustomGenesisFromPreState converts Ethereum test pre-state to Network CustomGenesisState
func (t *BlockTest) createCustomGenesisFromPreState(preState types.GenesisAlloc) network.CustomGenesisState {
	customGen := network.CustomGenesisState{}

	// Create genesis accounts and balances from preState
	var genesisAccounts []authtypes.GenesisAccount
	var balances []banktypes.Balance
	var evmAccounts []evmtypes.GenesisAccount

	for addr, account := range preState {
		// Convert Ethereum address to Cosmos address format
		cosmosAddr := sdk.AccAddress(addr.Bytes())

		// Create cosmos auth account with exact nonce
		baseAccount := authtypes.NewBaseAccount(
			cosmosAddr,
			nil, // pubkey will be set when account sends first tx
			0,   // account number
			account.Nonce,
		)
		genesisAccounts = append(genesisAccounts, baseAccount)

		// Create balance entry with exact balance from preState
		if account.Balance != nil && account.Balance.Sign() > 0 {
			balance := banktypes.Balance{
				Address: cosmosAddr.String(),
				Coins: sdk.NewCoins(sdk.NewCoin(
					sdk.DefaultBondDenom, // Use standard denomination
					cosmosmath.NewIntFromBigInt(account.Balance),
				)),
			}
			balances = append(balances, balance)
		}

		// Create EVM genesis account if it has code or storage
		if len(account.Code) > 0 || len(account.Storage) > 0 {
			var storage evmtypes.Storage
			for key, value := range account.Storage {
				storage = append(storage, evmtypes.State{
					Key:   key.Hex(),
					Value: value.Hex(),
				})
			}

			evmAccount := evmtypes.GenesisAccount{
				Address: addr.Hex(),
				Code:    hex.EncodeToString(account.Code),
				Storage: storage,
			}
			evmAccounts = append(evmAccounts, evmAccount)
		}
	}

	// Set up custom genesis state for auth module (accounts)
	if len(genesisAccounts) > 0 {
		authGenState := authtypes.DefaultGenesisState()
		for _, genesisAccount := range genesisAccounts {
			// Pack genesis account into Any type as required
			accAny, err := codectypes.NewAnyWithValue(genesisAccount)
			if err != nil {
				panic(fmt.Sprintf("failed to pack genesis account: %v", err))
			}
			authGenState.Accounts = append(authGenState.Accounts, accAny)
		}
		customGen[authtypes.ModuleName] = authGenState
	}

	// Set up custom genesis state for bank module (balances)
	if len(balances) > 0 {
		bankGenState := banktypes.DefaultGenesisState()
		bankGenState.Balances = balances
		customGen[banktypes.ModuleName] = bankGenState
	}

	// Set up custom genesis state for EVM module (contracts with code and storage)
	if len(evmAccounts) > 0 {
		evmGenState := evmtypes.DefaultGenesisState()
		evmGenState.Accounts = evmAccounts
		customGen[evmtypes.ModuleName] = evmGenState
	}

	return customGen
}

// executeBlocksWithSuite executes test blocks using the BlockTestSuite and Network
func (t *BlockTest) executeBlocksWithSuite(suite *BlockTestSuite, btJSON btJSON) error {
	ctx := suite.network.GetContext()

	// Get the underlying app through the network interface
	unitNetwork, ok := suite.network.(*network.UnitTestNetwork)
	if !ok {
		return fmt.Errorf("expected UnitTestNetwork, got %T", suite.network)
	}

	// Process each block
	for blockIndex, blockDef := range btJSON.Blocks {
		// Skip invalid blocks
		if blockDef.BlockHeader == nil {
			continue
		}

		// Decode the block to get transactions
		gethBlock, err := blockDef.decode()
		if err != nil {
			if blockDef.ExpectException != "" {
				continue
			}
			return fmt.Errorf("failed to decode block %d: %v", blockIndex, err)
		}

		// Execute transactions in the block
		for txIndex, tx := range gethBlock.Transactions() {
			// Convert Ethereum transaction to cosmos/evm message
			msgEthTx, err := t.convertToMsgEthereumTx(tx)
			if err != nil {
				return fmt.Errorf("failed to convert tx %d in block %d: %v", txIndex, blockIndex, err)
			}

			// Execute the transaction using the network's EVM capabilities
			evmKeeper := unitNetwork.App.GetEVMKeeper()
			_, err = evmKeeper.ApplyTransaction(ctx, msgEthTx)
			if err != nil {
				// Check if this failure was expected
				if blockDef.ExpectException != "" {
					continue
				}
				return fmt.Errorf("transaction execution failed: %v", err)
			}
		}
	}

	return nil
}

// validatePostStateWithSuite validates final state using the BlockTestSuite and Network
func (t *BlockTest) validatePostStateWithSuite(suite *BlockTestSuite, expectedPostState types.GenesisAlloc) error {
	ctx := suite.network.GetContext()

	// Get the underlying app through the network interface
	unitNetwork, ok := suite.network.(*network.UnitTestNetwork)
	if !ok {
		return fmt.Errorf("expected UnitTestNetwork, got %T", suite.network)
	}

	evmKeeper := unitNetwork.App.GetEVMKeeper()

	// Validate each expected account
	for addr, expectedAccount := range expectedPostState {
		// Get account balance using network's denom
		cosmosAddr := sdk.AccAddress(addr.Bytes())
		balance := unitNetwork.App.GetBankKeeper().GetBalance(ctx, cosmosAddr, suite.network.GetBaseDenom())
		actualBalance := balance.Amount.BigInt()

		if expectedAccount.Balance.Cmp(actualBalance) != 0 {
			return fmt.Errorf("balance mismatch for %s: expected %s, got %s",
				addr.Hex(), expectedAccount.Balance.String(), actualBalance.String())
		}

		// Get account nonce
		account := unitNetwork.App.GetAccountKeeper().GetAccount(ctx, cosmosAddr)
		if account == nil {
			if expectedAccount.Nonce != 0 {
				return fmt.Errorf("account %s not found, but expected nonce %d", addr.Hex(), expectedAccount.Nonce)
			}
		} else {
			if account.GetSequence() != expectedAccount.Nonce {
				return fmt.Errorf("nonce mismatch for %s: expected %d, got %d",
					addr.Hex(), expectedAccount.Nonce, account.GetSequence())
			}
		}

		// Get account code (if any)
		codeHash := evmKeeper.GetCodeHash(ctx, addr)
		code := evmKeeper.GetCode(ctx, codeHash)
		if !bytes.Equal(code, expectedAccount.Code) {
			return fmt.Errorf("code mismatch for %s", addr.Hex())
		}

		// Validate storage (if any)
		for key, expectedValue := range expectedAccount.Storage {
			actualValue := evmKeeper.GetState(ctx, addr, key)
			if actualValue != expectedValue {
				return fmt.Errorf("storage mismatch for %s at key %s: expected %s, got %s",
					addr.Hex(), key.Hex(), expectedValue.Hex(), actualValue.Hex())
			}
		}
	}

	return nil
}
