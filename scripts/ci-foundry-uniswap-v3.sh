#!/usr/bin/env bash

# CI script for running foundry-uniswap-v3 tests
# This script sets up dependencies, submodules, and runs the required forge script commands
# Usage: ./ci-foundry-uniswap-v3.sh [--verbose] [--node-log-print]

set -eo pipefail

VERBOSE=false
NODE_LOG_PRINT=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	--verbose | -v)
		VERBOSE=true
		shift
		;;
	--node-log-print)
		NODE_LOG_PRINT=true
		shift
		;;
	*)
		echo "Unknown option: $1"
		echo "Usage: $0 [--verbose] [--node-log-print]"
		exit 1
		;;
	esac
done

ROOT="$(git rev-parse --show-toplevel)"
TEST_DIR="$ROOT/tests/evm-tools-compatibility/foundry-uniswap-v3"

echo "Setting up foundry-uniswap-v3 tests..."

# Setup dependencies and submodules
echo "Running setup-compatibility-tests.sh..."
if [ "$NODE_LOG_PRINT" = true ]; then
	"$ROOT/scripts/setup-compatibility-tests.sh"
else
	"$ROOT/scripts/setup-compatibility-tests.sh" >/tmp/setup-compatibility-tests.log 2>&1
fi

# Launch evmd node
echo "Starting evmd node..."
pushd "$ROOT" >/dev/null
if [ "$NODE_LOG_PRINT" = true ]; then
	./local_node.sh -y &
else
	./local_node.sh -y >/tmp/evmd.log 2>&1 &
fi
NODE_PID=$!
popd >/dev/null

# Cleanup function to kill the node on exit
cleanup() {
	if [ -n "$NODE_PID" ]; then
		echo "Stopping evmd node..."
		kill "$NODE_PID" 2>/dev/null || true
		wait "$NODE_PID" 2>/dev/null || true
	fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Give the node a moment to start before checking
echo "Giving node time to initialize..."
sleep 3

# Wait for the node to be ready
echo "Waiting for evmd node to be ready..."
RPC_URL="http://127.0.0.1:8545"
TIMEOUT=60
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
	# Get the block number from the RPC endpoint
	RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
		--data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
		"$RPC_URL" 2>/dev/null || true)

	if [ -n "$RESPONSE" ]; then
		# Extract the hex block number from the JSON response
		BLOCK_HEX=$(echo "$RESPONSE" | grep -o '"result":"[^"]*"' | cut -d'"' -f4 || true)

		if [ -n "$BLOCK_HEX" ] && [ "$BLOCK_HEX" != "null" ]; then
			# Convert hex to decimal (handle potential errors)
			if BLOCK_NUMBER=$((16#${BLOCK_HEX#0x})); then
				echo "Current block number: $BLOCK_NUMBER (waiting for >= 5)"

				# Check if block number is >= 5
				if [ "$BLOCK_NUMBER" -ge 5 ]; then
					echo "Node is ready! Block number: $BLOCK_NUMBER"
					break
				fi
			fi
		fi
	fi

	echo "Waiting for node... ($ELAPSED/$TIMEOUT seconds)"

	sleep 2
	ELAPSED=$((ELAPSED + 2))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
	echo "Error: Node failed to reach block 5 within $TIMEOUT seconds"
	echo "Last response: $RESPONSE"
	echo "Checking node logs:"
	tail -20 /tmp/evmd.log 2>/dev/null || echo "No evmd logs found"
	exit 1
fi

# Change to the test directory
cd "$TEST_DIR"

# Source the environment file
if [ -f ".env" ]; then
	echo "Sourcing .env file..."
	# shellcheck source=/dev/null
	source .env
else
	echo "Error: No .env file found in $TEST_DIR"
	exit 1
fi

# Verify required environment variables are set
if [ -z "$CUSTOM_RPC" ] || [ -z "$CHAIN_ID" ] || [ -z "$LIBRARY_CONTRACT" ]; then
	echo "Error: Required environment variables not set"
	echo "CUSTOM_RPC: $CUSTOM_RPC"
	echo "CHAIN_ID: $CHAIN_ID"
	echo "LIBRARY_CONTRACT: $LIBRARY_CONTRACT"
	exit 1
fi

echo "Running foundry-uniswap-v3 deployment scripts..."

# Deploy NFTDescriptor
echo "Deploying NFTDescriptor..."
NFT_LOG_FILE="/tmp/nft_deployment.log"

# Run forge and tee output to both stdout and log file
if [ "$VERBOSE" = true ]; then
	forge script script/DeployNFTDescriptor.s.sol:DeployNFTDescriptor \
		--rpc-url "$CUSTOM_RPC" \
		--broadcast \
		--chain-id "$CHAIN_ID" 2>&1 | tee "$NFT_LOG_FILE"
else
	forge script script/DeployNFTDescriptor.s.sol:DeployNFTDescriptor \
		--rpc-url "$CUSTOM_RPC" \
		--broadcast \
		--chain-id "$CHAIN_ID" > "$NFT_LOG_FILE" 2>&1
fi

NFT_EXIT_CODE=${PIPESTATUS[0]}

# Give a moment for output to be fully written to log file
sleep 2

# Check for success
if [ $NFT_EXIT_CODE -ne 0 ]; then
	echo "Error: NFTDescriptor deployment failed with exit code $NFT_EXIT_CODE"
	echo "Last 20 lines of output:"
	tail -20 "$NFT_LOG_FILE"
	exit 1
fi

if ! grep -q "ONCHAIN EXECUTION COMPLETE & SUCCESSFUL" "$NFT_LOG_FILE"; then
	echo "Error: NFTDescriptor deployment did not complete successfully"
	echo "Last 20 lines of output:"
	tail -20 "$NFT_LOG_FILE"
	exit 1
fi

echo "NFTDescriptor deployment completed successfully!"

# Take a rest to ensure the deployment is complete
echo "Waiting for NFTDescriptor deployment to complete..."
sleep 5

# Deploy UniswapV3 with NFTDescriptor library
echo "Deploying UniswapV3..."
UNISWAP_LOG_FILE="/tmp/uniswap_deployment.log"

# Run forge and tee output to both stdout and log file
if [ "$VERBOSE" = true ]; then
	forge script script/DeployUniswapV3.s.sol:DeployUniswapV3 \
		--rpc-url "$CUSTOM_RPC" \
		--chain-id "$CHAIN_ID" \
		--broadcast \
		--slow \
		--private-key "$PRIVATE_KEY" \
		--libraries "lib/v3-periphery/contracts/libraries/NFTDescriptor.sol:NFTDescriptor:$LIBRARY_CONTRACT" 2>&1 | tee "$UNISWAP_LOG_FILE"
else
	forge script script/DeployUniswapV3.s.sol:DeployUniswapV3 \
		--rpc-url "$CUSTOM_RPC" \
		--chain-id "$CHAIN_ID" \
		--broadcast \
		--slow \
		--private-key "$PRIVATE_KEY" \
		--libraries "lib/v3-periphery/contracts/libraries/NFTDescriptor.sol:NFTDescriptor:$LIBRARY_CONTRACT" > "$UNISWAP_LOG_FILE" 2>&1
fi

UNISWAP_EXIT_CODE=${PIPESTATUS[0]}

# Give a moment for output to be fully written to log file
sleep 2

# Check for success
if [ $UNISWAP_EXIT_CODE -ne 0 ]; then
	echo "Error: UniswapV3 deployment failed with exit code $UNISWAP_EXIT_CODE"
	echo "Last 20 lines of output:"
	tail -20 "$UNISWAP_LOG_FILE"
	exit 1
fi

if ! grep -q "ONCHAIN EXECUTION COMPLETE & SUCCESSFUL" "$UNISWAP_LOG_FILE"; then
	echo "Error: UniswapV3 deployment did not complete successfully"
	echo "Last 20 lines of output:"
	tail -20 "$UNISWAP_LOG_FILE"
	exit 1
fi

echo "UniswapV3 deployment completed successfully!"
echo "foundry-uniswap-v3 tests completed successfully!"
