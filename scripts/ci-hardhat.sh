#!/usr/bin/env bash

# CI script for running hardhat compatibility tests
# This script sets up dependencies, launches the node, and runs hardhat tests
# Usage: ./ci-hardhat.sh [--verbose] [--node-log-print]

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
TEST_DIR="$ROOT/tests/evm-tools-compatibility/hardhat"

echo "Setting up hardhat compatibility tests..."

# Setup dependencies
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
				echo "Current block number: $BLOCK_NUMBER (waiting for >= 10)"

				# Check if block number is >= 10
				if [ "$BLOCK_NUMBER" -ge 10 ]; then
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
	echo "Error: Node failed to reach block 10 within $TIMEOUT seconds"
	echo "Last response: $RESPONSE"
	echo "Checking node logs:"
	tail -20 /tmp/evmd.log 2>/dev/null || echo "No evmd logs found"
	exit 1
fi

# Change to the test directory
cd "$TEST_DIR"

# Install npm dependencies if not already installed
if [ ! -d "node_modules" ]; then
	echo "Installing npm dependencies..."
	npm install
else
	echo "npm dependencies already installed, skipping..."
fi

echo "Running hardhat compatibility tests..."

# Run tests with npx hardhat test (default network)
if [ "$VERBOSE" = true ]; then
	echo "Running: npx hardhat test"
	npx hardhat test 2>&1 | tee /tmp/hardhat-test.log
else
	echo "Running: npx hardhat test"
	npx hardhat test 2>&1 | tee /tmp/hardhat-test.log
fi

# Check if tests passed and no failures occurred
if [ "${PIPESTATUS[0]}" -eq 0 ] && ! grep -i "failing" /tmp/hardhat-test.log >/dev/null; then
	echo "All hardhat compatibility tests (default network) passed successfully!"
else
	echo "Error: Some hardhat tests (default network) failed"
	echo "Test output:"
	tail -20 /tmp/hardhat-test.log
	if grep -i "failing" /tmp/hardhat-test.log >/dev/null; then
		echo "Found 'failing' keyword in test output"
	fi
	exit 1
fi

echo "Running hardhat compatibility tests with localhost network..."

# Run tests with npx hardhat test --network localhost
if [ "$VERBOSE" = true ]; then
	echo "Running: npx hardhat test --network localhost"
	npx hardhat test --network localhost 2>&1 | tee /tmp/hardhat-test-localhost.log
else
	echo "Running: npx hardhat test --network localhost"
	npx hardhat test --network localhost 2>&1 | tee /tmp/hardhat-test-localhost.log
fi

# Check if tests passed and no failures occurred
if [ "${PIPESTATUS[0]}" -eq 0 ] && ! grep -i "failing" /tmp/hardhat-test-localhost.log >/dev/null; then
	echo "All hardhat compatibility tests (localhost network) passed successfully!"
else
	echo "Error: Some hardhat tests (localhost network) failed"
	echo "Test output:"
	tail -20 /tmp/hardhat-test-localhost.log
	if grep -i "failing" /tmp/hardhat-test-localhost.log >/dev/null; then
		echo "Found 'failing' keyword in test output"
	fi
	exit 1
fi
