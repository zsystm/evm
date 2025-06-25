package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/cosmos/evm/rpc/backend"
	rpctypes "github.com/cosmos/evm/rpc/types"
	"github.com/cosmos/evm/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"cosmossdk.io/log"
)

// The Ethereum API allows applications to connect to an node of any Cosmos EVM based blockchain.
// Developers can interact with on-chain EVM data
// and send different types of transactions to the network by utilizing the
// endpoints provided by the API. The API follows a JSON-RPC standard. If not
// otherwise specified, the interface is derived from the Alchemy Ethereum API:
// https://docs.alchemy.com/alchemy/apis/ethereum
type EthereumAPI interface {
	// Getting Blocks
	//
	// Retrieves information from a particular block in the blockchain.
	BlockNumber() (hexutil.Uint64, error)
	GetBlockByNumber(ethBlockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint
	GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint

	// Reading Transactions
	//
	// Retrieves information on the state data for addresses regardless of whether
	// it is a user or a smart contract.
	GetTransactionByHash(hash common.Hash) (*rpctypes.RPCTransaction, error)
	GetTransactionCount(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Uint64, error)
	GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error)
	GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error)
	GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error)
	// eth_getBlockReceipts

	// Writing Transactions
	//
	// Allows developers to both send ETH from one address to another, write data
	// on-chain, and interact with smart contracts.
	SendRawTransaction(data hexutil.Bytes) (common.Hash, error)
	SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error)
	// eth_sendPrivateTransaction
	// eth_cancel	PrivateTransaction

	// Account Information
	//
	// Returns information regarding an address's stored on-chain data.
	Accounts() ([]common.Address, error)
	GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error)
	GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error)
	GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error)
	GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error)

	// EVM/Smart Contract Execution
	//
	// Allows developers to read data from the blockchain which includes executing
	// smart contracts. However, no data is published to the Ethereum network.
	Call(args evmtypes.TransactionArgs, blockNrOrHash rpctypes.BlockNumberOrHash, _ *rpctypes.StateOverride) (hexutil.Bytes, error)

	// Chain Information
	//
	// Returns information on the Ethereum network and internal settings.
	ProtocolVersion() hexutil.Uint
	GasPrice() (*hexutil.Big, error)
	EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error)
	FeeHistory(blockCount, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (*rpctypes.FeeHistoryResult, error)
	MaxPriorityFeePerGas() (*hexutil.Big, error)
	ChainId() (*hexutil.Big, error)

	// Getting Uncles
	//
	// Returns information on uncle blocks are which are network rejected blocks and replaced by a canonical block instead.
	GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{}
	GetUncleByBlockNumberAndIndex(number, idx hexutil.Uint) map[string]interface{}
	GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint
	GetUncleCountByBlockNumber(blockNum rpctypes.BlockNumber) hexutil.Uint

	// Proof of Work
	Hashrate() hexutil.Uint64
	Mining() bool

	// Other
	Syncing() (interface{}, error)
	Coinbase() (string, error)
	Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error)
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	SignTypedData(address common.Address, typedData apitypes.TypedData) (hexutil.Bytes, error)
	FillTransaction(args evmtypes.TransactionArgs) (*rpctypes.SignTransactionResult, error)
	Resend(ctx context.Context, args evmtypes.TransactionArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error)
	GetPendingTransactions() ([]*rpctypes.RPCTransaction, error)
	// eth_signTransaction (on Ethereum.org)
	// eth_getCompilers (on Ethereum.org)
	// eth_compileSolidity (on Ethereum.org)
	// eth_compileLLL (on Ethereum.org)
	// eth_compileSerpent (on Ethereum.org)
	// eth_getWork (on Ethereum.org)
	// eth_submitWork (on Ethereum.org)
	// eth_submitHashrate (on Ethereum.org)
}

var _ EthereumAPI = (*PublicAPI)(nil)

// PublicAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicAPI struct {
	ctx     context.Context
	logger  log.Logger
	backend backend.EVMBackend
}

// NewPublicAPI creates an instance of the public ETH Web3 API.
func NewPublicAPI(logger log.Logger, backend backend.EVMBackend) *PublicAPI {
	api := &PublicAPI{
		ctx:     context.Background(),
		logger:  logger.With("client", "json-rpc"),
		backend: backend,
	}

	return api
}

///////////////////////////////////////////////////////////////////////////////
///                           Blocks						                            ///
///////////////////////////////////////////////////////////////////////////////

// BlockNumber returns the current block number.
func (e *PublicAPI) BlockNumber() (hexutil.Uint64, error) {
	e.logger.Debug("eth_blockNumber")
	return e.backend.BlockNumber()
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicAPI) GetBlockByNumber(ethBlockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debug("eth_getBlockByNumber", "number", ethBlockNum, "full", fullTx)
	return e.backend.GetBlockByNumber(ethBlockNum, fullTx)
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debug("eth_getBlockByHash", "hash", hash.Hex(), "full", fullTx)
	return e.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockReceipts returns the block receipts for the given block hash or number or tag.
func (e *PublicAPI) GetBlockReceipts(ctx context.Context, blockNrOrHash rpctypes.BlockNumberOrHash) ([]map[string]interface{}, error) {
	e.logger.Debug("eth_getBlockReceipts", "block number or hash", blockNrOrHash)
	return e.backend.GetBlockReceipts(blockNrOrHash)
}

///////////////////////////////////////////////////////////////////////////////
///                           Read Txs					                            ///
///////////////////////////////////////////////////////////////////////////////

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicAPI) GetTransactionByHash(hash common.Hash) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByHash", "hash", hash.Hex())
	return e.backend.GetTransactionByHash(hash)
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicAPI) GetTransactionCount(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Uint64, error) {
	e.logger.Debug("eth_getTransactionCount", "address", address.Hex(), "block number or hash", blockNrOrHash)
	blockNum, err := e.backend.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	return e.backend.GetTransactionCount(address, blockNum)
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	hexTx := hash.Hex()
	e.logger.Debug("eth_getTransactionReceipt", "hash", hexTx)
	return e.backend.GetTransactionReceipt(hash)
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	e.logger.Debug("eth_getBlockTransactionCountByHash", "hash", hash.Hex())
	return e.backend.GetBlockTransactionCountByHash(hash)
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	e.logger.Debug("eth_getBlockTransactionCountByNumber", "height", blockNum.Int64())
	return e.backend.GetBlockTransactionCountByNumber(blockNum)
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByBlockHashAndIndex", "hash", hash.Hex(), "index", idx)
	return e.backend.GetTransactionByBlockHashAndIndex(hash, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)
	return e.backend.GetTransactionByBlockNumberAndIndex(blockNum, idx)
}

///////////////////////////////////////////////////////////////////////////////
///                           Write Txs					                            ///
///////////////////////////////////////////////////////////////////////////////

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	e.logger.Debug("eth_sendRawTransaction", "length", len(data))
	return e.backend.SendRawTransaction(data)
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicAPI) SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error) {
	e.logger.Debug("eth_sendTransaction", "args", args.String())
	return e.backend.SendTransaction(args)
}

///////////////////////////////////////////////////////////////////////////////
///                           Account Information				                    ///
///////////////////////////////////////////////////////////////////////////////

// Accounts returns the list of accounts available to this node.
func (e *PublicAPI) Accounts() ([]common.Address, error) {
	e.logger.Debug("eth_accounts")
	return e.backend.Accounts()
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicAPI) GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
	e.logger.Debug("eth_getBalance", "address", address.String(), "block number or hash", blockNrOrHash)
	return e.backend.GetBalance(address, blockNrOrHash)
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicAPI) GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	e.logger.Debug("eth_getStorageAt", "address", address.Hex(), "key", key, "block number or hash", blockNrOrHash)
	return e.backend.GetStorageAt(address, key, blockNrOrHash)
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicAPI) GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	e.logger.Debug("eth_getCode", "address", address.Hex(), "block number or hash", blockNrOrHash)
	// if address is same with 0x800, then return the mock code
	bytecode := "608060405234801561001057600080fd5b50600436106100b45760003560e01c806353266bbb1161007157806353266bbb146101dc57806354b826f51461020c5780637d9f939c1461023c578063a03ffee11461026c578063a50f05ac1461029c578063f7cd5516146102cc576100b4565b806310a2851c146100b957806312d58dfe146100ea578063186b21671461011a578063223b3b7a1461014b578063241774e61461017b5780633edab33c146101ac575b600080fd5b6100d360048036038101906100ce9190610835565b6102fc565b6040516100e1929190610dd2565b60405180910390f35b61010460048036038101906100ff9190610e35565b610311565b6040516101119190610ed8565b60405180910390f35b610134600480360381019061012f9190610ef3565b610387565b6040516101429291906111b4565b60405180910390f35b610165600480360381019061016091906111eb565b610399565b6040516101729190611315565b60405180910390f35b61019560048036038101906101909190611337565b6103a6565b6040516101a39291906113e3565b60405180910390f35b6101c660048036038101906101c19190611413565b6103e3565b6040516101d39190611496565b60405180910390f35b6101f660048036038101906101f191906115e1565b61045c565b6040516102039190610ed8565b60405180910390f35b61022660048036038101906102219190611650565b6104d0565b6040516102339190611496565b60405180910390f35b610256600480360381019061025191906116f7565b610562565b6040516102639190611804565b60405180910390f35b61028660048036038101906102819190611337565b610573565b60405161029391906119ae565b60405180910390f35b6102b660048036038101906102b19190611a25565b610582565b6040516102c39190610ed8565b60405180910390f35b6102e660048036038101906102e19190611ac7565b6105e0565b6040516102f39190610ed8565b60405180910390f35b606061030661063f565b965096945050505050565b60008573ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff167f6dbe2fb6b2613bdd8e3d284a6111592e06c3ab0af846ff89b6688d48f408dbb58585604051610372929190611b93565b60405180910390a36001905095945050505050565b606061039161063f565b935093915050565b6103a1610663565b919050565b60006103b06106d7565b60405180604001604052806040518060200160405280600081525081526020016000815250905060009150935093915050565b6000804290508573ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff167f4bf8087be3b8a59c2662514df2ed4a3dcaf9ca22f442340cfc05a4e52343d18e8584604051610448929190611b93565b60405180910390a380915050949350505050565b60008373ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f500599802164a08023e87ffc3eed0ba3ae60697b3083ba81d046683679d81c6b84856040516104bd929190611b93565b60405180910390a3600190509392505050565b6000804290508773ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f82b07f2421474f1e3f1e0b34738cb5ffb925273f408e7591d9c803dcae8da657868560405161054c929190611b93565b60405180910390a4809150509695505050505050565b61056a6106f1565b95945050505050565b61057b610719565b9392505050565b60008373ffffffffffffffffffffffffffffffffffffffff167fdce27cf2792bd8d8f28df5d2cdf379cd593414f21332370ca808c1e703eb4e1f84846040516105cc929190611bcb565b60405180910390a260019050949350505050565b60008473ffffffffffffffffffffffffffffffffffffffff167f9bdb560f8135cb46033a55410c14e14b1a7bc2d3f3e9973f4b49533e176468b0836040516106289190611bf4565b60405180910390a260019050979650505050505050565b604051806040016040528060608152602001600067ffffffffffffffff1681525090565b60405180610160016040528060608152602001606081526020016000151581526020016000600381111561069a57610699610faa565b5b8152602001600081526020016000815260200160608152602001600060070b8152602001600060070b815260200160008152602001600081525090565b604051806040016040528060608152602001600081525090565b6040518060800160405280606081526020016060815260200160608152602001606081525090565b60405180606001604052806060815260200160608152602001606081525090565b6000604051905090565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006107798261074e565b9050919050565b6107898161076e565b811461079457600080fd5b50565b6000813590506107a681610780565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126107d1576107d06107ac565b5b8235905067ffffffffffffffff8111156107ee576107ed6107b1565b5b60208301915083600182028301111561080a576108096107b6565b5b9250929050565b600080fd5b600060a0828403121561082c5761082b610811565b5b81905092915050565b6000806000806000806080878903121561085257610851610744565b5b600061086089828a01610797565b965050602087013567ffffffffffffffff81111561088157610880610749565b5b61088d89828a016107bb565b9550955050604087013567ffffffffffffffff8111156108b0576108af610749565b5b6108bc89828a016107bb565b9350935050606087013567ffffffffffffffff8111156108df576108de610749565b5b6108eb89828a01610816565b9150509295509295509295565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561095e578082015181840152602081019050610943565b60008484015250505050565b6000601f19601f8301169050919050565b600061098682610924565b610990818561092f565b93506109a0818560208601610940565b6109a98161096a565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b60008160070b9050919050565b6109f6816109e0565b82525050565b6000819050919050565b610a0f816109fc565b82525050565b608082016000820151610a2b60008501826109ed565b506020820151610a3e60208501826109ed565b506040820151610a516040850182610a06565b506060820151610a646060850182610a06565b50505050565b6000610a768383610a15565b60808301905092915050565b6000602082019050919050565b6000610a9a826109b4565b610aa481856109bf565b9350610aaf836109d0565b8060005b83811015610ae0578151610ac78882610a6a565b9750610ad283610a82565b925050600181019050610ab3565b5085935050505092915050565b60006080830160008301518482036000860152610b0a828261097b565b91505060208301518482036020860152610b24828261097b565b91505060408301518482036040860152610b3e828261097b565b91505060608301518482036060860152610b588282610a8f565b9150508091505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b60a082016000820151610ba76000850182610a15565b506020820151610bba6080850182610a06565b50505050565b6000610bcc8383610b91565b60a08301905092915050565b6000602082019050919050565b6000610bf082610b65565b610bfa8185610b70565b9350610c0583610b81565b8060005b83811015610c36578151610c1d8882610bc0565b9750610c2883610bd8565b925050600181019050610c09565b5085935050505092915050565b60006040830160008301518482036000860152610c608282610aed565b91505060208301518482036020860152610c7a8282610be5565b9150508091505092915050565b6000610c938383610c43565b905092915050565b6000602082019050919050565b6000610cb3826108f8565b610cbd8185610903565b935083602082028501610ccf85610914565b8060005b85811015610d0b5784840389528151610cec8582610c87565b9450610cf783610c9b565b925060208a01995050600181019050610cd3565b50829750879550505050505092915050565b600081519050919050565b600082825260208201905092915050565b6000610d4482610d1d565b610d4e8185610d28565b9350610d5e818560208601610940565b610d678161096a565b840191505092915050565b600067ffffffffffffffff82169050919050565b610d8f81610d72565b82525050565b60006040830160008301518482036000860152610db28282610d39565b9150506020830151610dc76020860182610d86565b508091505092915050565b60006040820190508181036000830152610dec8185610ca8565b90508181036020830152610e008184610d95565b90509392505050565b610e12816109fc565b8114610e1d57600080fd5b50565b600081359050610e2f81610e09565b92915050565b600080600080600060808688031215610e5157610e50610744565b5b6000610e5f88828901610797565b955050602086013567ffffffffffffffff811115610e8057610e7f610749565b5b610e8c888289016107bb565b94509450506040610e9f88828901610e20565b9250506060610eb088828901610e20565b9150509295509295909350565b60008115159050919050565b610ed281610ebd565b82525050565b6000602082019050610eed6000830184610ec9565b92915050565b600080600060408486031215610f0c57610f0b610744565b5b600084013567ffffffffffffffff811115610f2a57610f29610749565b5b610f36868287016107bb565b9350935050602084013567ffffffffffffffff811115610f5957610f58610749565b5b610f6586828701610816565b9150509250925092565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b610fa481610ebd565b82525050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b60048110610fea57610fe9610faa565b5b50565b6000819050610ffb82610fd9565b919050565b600061100b82610fed565b9050919050565b61101b81611000565b82525050565b600061016083016000830151848203600086015261103f828261097b565b91505060208301518482036020860152611059828261097b565b915050604083015161106e6040860182610f9b565b5060608301516110816060860182611012565b5060808301516110946080860182610a06565b5060a08301516110a760a0860182610a06565b5060c083015184820360c08601526110bf828261097b565b91505060e08301516110d460e08601826109ed565b506101008301516110e96101008601826109ed565b506101208301516110fe610120860182610a06565b50610140830151611113610140860182610a06565b508091505092915050565b600061112a8383611021565b905092915050565b6000602082019050919050565b600061114a82610f6f565b6111548185610f7a565b93508360208202850161116685610f8b565b8060005b858110156111a25784840389528151611183858261111e565b945061118e83611132565b925060208a0199505060018101905061116a565b50829750879550505050505092915050565b600060408201905081810360008301526111ce818561113f565b905081810360208301526111e28184610d95565b90509392505050565b60006020828403121561120157611200610744565b5b600061120f84828501610797565b91505092915050565b6000610160830160008301518482036000860152611236828261097b565b91505060208301518482036020860152611250828261097b565b91505060408301516112656040860182610f9b565b5060608301516112786060860182611012565b50608083015161128b6080860182610a06565b5060a083015161129e60a0860182610a06565b5060c083015184820360c08601526112b6828261097b565b91505060e08301516112cb60e08601826109ed565b506101008301516112e06101008601826109ed565b506101208301516112f5610120860182610a06565b5061014083015161130a610140860182610a06565b508091505092915050565b6000602082019050818103600083015261132f8184611218565b905092915050565b6000806000604084860312156113505761134f610744565b5b600061135e86828701610797565b935050602084013567ffffffffffffffff81111561137f5761137e610749565b5b61138b868287016107bb565b92509250509250925092565b6113a0816109fc565b82525050565b600060408301600083015184820360008601526113c3828261097b565b91505060208301516113d86020860182610a06565b508091505092915050565b60006040820190506113f86000830185611397565b818103602083015261140a81846113a6565b90509392505050565b6000806000806060858703121561142d5761142c610744565b5b600061143b87828801610797565b945050602085013567ffffffffffffffff81111561145c5761145b610749565b5b611468878288016107bb565b9350935050604061147b87828801610e20565b91505092959194509250565b611490816109e0565b82525050565b60006020820190506114ab6000830184611487565b92915050565b600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6114ee8261096a565b810181811067ffffffffffffffff8211171561150d5761150c6114b6565b5b80604052505050565b600061152061073a565b905061152c82826114e5565b919050565b600067ffffffffffffffff82111561154c5761154b6114b6565b5b6115558261096a565b9050602081019050919050565b82818337600083830152505050565b600061158461157f84611531565b611516565b9050828152602081018484840111156115a05761159f6114b1565b5b6115ab848285611562565b509392505050565b600082601f8301126115c8576115c76107ac565b5b81356115d8848260208601611571565b91505092915050565b6000806000606084860312156115fa576115f9610744565b5b600061160886828701610797565b935050602084013567ffffffffffffffff81111561162957611628610749565b5b611635868287016115b3565b925050604061164686828701610e20565b9150509250925092565b6000806000806000806080878903121561166d5761166c610744565b5b600061167b89828a01610797565b965050602087013567ffffffffffffffff81111561169c5761169b610749565b5b6116a889828a016107bb565b9550955050604087013567ffffffffffffffff8111156116cb576116ca610749565b5b6116d789828a016107bb565b935093505060606116ea89828a01610e20565b9150509295509295509295565b60008060008060006060868803121561171357611712610744565b5b600061172188828901610797565b955050602086013567ffffffffffffffff81111561174257611741610749565b5b61174e888289016107bb565b9450945050604086013567ffffffffffffffff81111561177157611770610749565b5b61177d888289016107bb565b92509250509295509295909350565b600060808301600083015184820360008601526117a9828261097b565b915050602083015184820360208601526117c3828261097b565b915050604083015184820360408601526117dd828261097b565b915050606083015184820360608601526117f78282610a8f565b9150508091505092915050565b6000602082019050818103600083015261181e818461178c565b905092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b60c08201600082015161186860008501826109ed565b50602082015161187b60208501826109ed565b50604082015161188e6040850182610a06565b5060608201516118a16060850182610a06565b5060808201516118b46080850182610d86565b5060a08201516118c760a08501826109ed565b50505050565b60006118d98383611852565b60c08301905092915050565b6000602082019050919050565b60006118fd82611826565b6119078185611831565b935061191283611842565b8060005b8381101561194357815161192a88826118cd565b9750611935836118e5565b925050600181019050611916565b5085935050505092915050565b6000606083016000830151848203600086015261196d828261097b565b91505060208301518482036020860152611987828261097b565b915050604083015184820360408601526119a182826118f2565b9150508091505092915050565b600060208201905081810360008301526119c88184611950565b905092915050565b600060a082840312156119e6576119e5610811565b5b81905092915050565b6000819050919050565b611a02816119ef565b8114611a0d57600080fd5b50565b600081359050611a1f816119f9565b92915050565b60008060008060808587031215611a3f57611a3e610744565b5b600085013567ffffffffffffffff811115611a5d57611a5c610749565b5b611a69878288016119d0565b9450506020611a7a87828801610797565b9350506040611a8b87828801611a10565b9250506060611a9c87828801611a10565b91505092959194509250565b600060608284031215611abe57611abd610811565b5b81905092915050565b6000806000806000806000610100888a031215611ae757611ae6610744565b5b600088013567ffffffffffffffff811115611b0557611b04610749565b5b611b118a828b016119d0565b9750506020611b228a828b01611aa8565b9650506080611b338a828b01610e20565b95505060a0611b448a828b01610797565b94505060c088013567ffffffffffffffff811115611b6557611b64610749565b5b611b718a828b016107bb565b935093505060e0611b848a828b01610e20565b91505092959891949750929550565b6000604082019050611ba86000830185611397565b611bb56020830184611397565b9392505050565b611bc5816119ef565b82525050565b6000604082019050611be06000830185611bbc565b611bed6020830184611bbc565b9392505050565b6000602082019050611c096000830184611397565b9291505056fea26469706673582212202dd34fe8e56ebe9e67168a7b6ca8b8d3b50a3310dfcb21178a53e9ea36bce6d864736f6c634300081e0033"
	if address == common.HexToAddress("0x800") {
		return hexutil.Bytes(common.FromHex(bytecode)), nil
	}
	return e.backend.GetCode(address, blockNrOrHash)
}

// GetProof returns an account object with proof and any storage proofs
func (e *PublicAPI) GetProof(address common.Address,
	storageKeys []string,
	blockNrOrHash rpctypes.BlockNumberOrHash,
) (*rpctypes.AccountResult, error) {
	e.logger.Debug("eth_getProof", "address", address.Hex(), "keys", storageKeys, "block number or hash", blockNrOrHash)
	return e.backend.GetProof(address, storageKeys, blockNrOrHash)
}

///////////////////////////////////////////////////////////////////////////////
///                           EVM/Smart Contract Execution				          ///
///////////////////////////////////////////////////////////////////////////////

// Call performs a raw contract call.
func (e *PublicAPI) Call(args evmtypes.TransactionArgs,
	blockNrOrHash rpctypes.BlockNumberOrHash,
	_ *rpctypes.StateOverride,
) (hexutil.Bytes, error) {
	e.logger.Debug("eth_call", "args", args.String(), "block number or hash", blockNrOrHash)

	blockNum, err := e.backend.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	data, err := e.backend.DoCall(args, blockNum)
	if err != nil {
		return []byte{}, err
	}

	return (hexutil.Bytes)(data.Ret), nil
}

///////////////////////////////////////////////////////////////////////////////
///                           Event Logs													          ///
///////////////////////////////////////////////////////////////////////////////
// FILTER API at ./filters/api.go

///////////////////////////////////////////////////////////////////////////////
///                           Chain Information										          ///
///////////////////////////////////////////////////////////////////////////////

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicAPI) ProtocolVersion() hexutil.Uint {
	e.logger.Debug("eth_protocolVersion")
	return hexutil.Uint(types.ProtocolVersion)
}

// GasPrice returns the current gas price based on Cosmos EVM's gas price oracle.
func (e *PublicAPI) GasPrice() (*hexutil.Big, error) {
	e.logger.Debug("eth_gasPrice")
	return e.backend.GasPrice()
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (e *PublicAPI) EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error) {
	e.logger.Debug("eth_estimateGas")
	return e.backend.EstimateGas(args, blockNrOptional)
}

func (e *PublicAPI) FeeHistory(blockCount,
	lastBlock rpc.BlockNumber,
	rewardPercentiles []float64,
) (*rpctypes.FeeHistoryResult, error) {
	e.logger.Debug("eth_feeHistory")
	return e.backend.FeeHistory(blockCount, lastBlock, rewardPercentiles)
}

// MaxPriorityFeePerGas returns a suggestion for a gas tip cap for dynamic fee transactions.
func (e *PublicAPI) MaxPriorityFeePerGas() (*hexutil.Big, error) {
	e.logger.Debug("eth_maxPriorityFeePerGas")
	head, err := e.backend.CurrentHeader()
	if err != nil {
		return nil, err
	}
	tipcap, err := e.backend.SuggestGasTipCap(head.BaseFee)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(tipcap), nil
}

// ChainId is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (e *PublicAPI) ChainId() (*hexutil.Big, error) { //nolint
	e.logger.Debug("eth_chainId")
	return e.backend.ChainID()
}

///////////////////////////////////////////////////////////////////////////////
///                           Uncles															          ///
///////////////////////////////////////////////////////////////////////////////

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicAPI) GetUncleByBlockHashAndIndex(_ common.Hash, _ hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicAPI) GetUncleByBlockNumberAndIndex(_, _ hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleCountByBlockHash returns the number of uncles in the block identified by hash. Always zero.
func (e *PublicAPI) GetUncleCountByBlockHash(_ common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block identified by number. Always zero.
func (e *PublicAPI) GetUncleCountByBlockNumber(_ rpctypes.BlockNumber) hexutil.Uint {
	return 0
}

///////////////////////////////////////////////////////////////////////////////
///                           Proof of Work												          ///
///////////////////////////////////////////////////////////////////////////////

// Hashrate returns the current node's hashrate. Always 0.
func (e *PublicAPI) Hashrate() hexutil.Uint64 {
	e.logger.Debug("eth_hashrate")
	return 0
}

// Mining returns whether or not this node is currently mining. Always false.
func (e *PublicAPI) Mining() bool {
	e.logger.Debug("eth_mining")
	return false
}

///////////////////////////////////////////////////////////////////////////////
///                           Other 															          ///
///////////////////////////////////////////////////////////////////////////////

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronize from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (e *PublicAPI) Syncing() (interface{}, error) {
	e.logger.Debug("eth_syncing")
	return e.backend.Syncing()
}

// Coinbase is the address that staking rewards will be send to (alias for Etherbase).
func (e *PublicAPI) Coinbase() (string, error) {
	e.logger.Debug("eth_coinbase")

	coinbase, err := e.backend.GetCoinbase()
	if err != nil {
		return "", err
	}
	ethAddr := common.BytesToAddress(coinbase.Bytes())
	return ethAddr.Hex(), nil
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	e.logger.Debug("eth_sign", "address", address.Hex(), "data", common.Bytes2Hex(data))
	return e.backend.Sign(address, data)
}

// GetTransactionLogs returns the logs given a transaction hash.
func (e *PublicAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	e.logger.Debug("eth_getTransactionLogs", "hash", txHash)

	return e.backend.GetTransactionLogs(txHash)
}

// SignTypedData signs EIP-712 conformant typed data
func (e *PublicAPI) SignTypedData(address common.Address, typedData apitypes.TypedData) (hexutil.Bytes, error) {
	e.logger.Debug("eth_signTypedData", "address", address.Hex(), "data", typedData)
	return e.backend.SignTypedData(address, typedData)
}

// FillTransaction fills the defaults (nonce, gas, gasPrice or 1559 fields)
// on a given unsigned transaction, and returns it to the caller for further
// processing (signing + broadcast).
func (e *PublicAPI) FillTransaction(args evmtypes.TransactionArgs) (*rpctypes.SignTransactionResult, error) {
	// Set some sanity defaults and terminate on failure
	args, err := e.backend.SetTxDefaults(args)
	if err != nil {
		return nil, err
	}

	// Assemble the transaction and obtain rlp
	tx := args.ToTransaction().AsTransaction()

	data, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &rpctypes.SignTransactionResult{
		Raw: data,
		Tx:  tx,
	}, nil
}

// Resend accepts an existing transaction and a new gas price and limit. It will remove
// the given transaction from the pool and reinsert it with the new gas price and limit.
func (e *PublicAPI) Resend(_ context.Context,
	args evmtypes.TransactionArgs,
	gasPrice *hexutil.Big,
	gasLimit *hexutil.Uint64,
) (common.Hash, error) {
	e.logger.Debug("eth_resend", "args", args.String())
	return e.backend.Resend(args, gasPrice, gasLimit)
}

// GetPendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *PublicAPI) GetPendingTransactions() ([]*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getPendingTransactions")

	txs, err := e.backend.PendingTransactions()
	if err != nil {
		return nil, err
	}

	chainIDHex, err := e.backend.ChainID()
	if err != nil {
		return nil, err
	}

	chainID := chainIDHex.ToInt()

	result := make([]*rpctypes.RPCTransaction, 0, len(txs))
	for _, tx := range txs {
		for _, msg := range (*tx).GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				// not valid ethereum tx
				break
			}

			rpctx, err := rpctypes.NewTransactionFromMsg(
				ethMsg,
				common.Hash{},
				uint64(0),
				uint64(0),
				nil,
				chainID,
			)
			if err != nil {
				return nil, err
			}

			result = append(result, rpctx)
		}
	}

	return result, nil
}
