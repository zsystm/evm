#!/usr/bin/env bash

# CI script for running foundry-uniswap-v3 tests
# This script sets up dependencies, submodules, and runs the required forge script commands
# Usage: ./ci-foundry-uniswap-v3.sh [--verbose]

set -eo pipefail

VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--verbose]"
            exit 1
            ;;
    esac
done

ROOT="$(git rev-parse --show-toplevel)"
TEST_DIR="$ROOT/tests/evm-tools-compatibility/foundry-uniswap-v3"

echo "Setting up foundry-uniswap-v3 tests..."

# Setup dependencies and submodules
echo "Running setup-compatibility-tests.sh..."
if [ "$VERBOSE" = true ]; then
    "$ROOT/scripts/setup-compatibility-tests.sh"
else
    "$ROOT/scripts/setup-compatibility-tests.sh" > /tmp/setup-compatibility-tests.log 2>&1
fi

# Launch evmd node
echo "Starting evmd node..."
pushd "$ROOT" >/dev/null
if [ "$VERBOSE" = true ]; then
    ./local_node.sh -y --no-install &
else
    ./local_node.sh -y --no-install > /tmp/evmd.log 2>&1 &
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
            if BLOCK_NUMBER=$((16#${BLOCK_HEX#0x} 2>/dev/null)); then
                if [ "$VERBOSE" = true ]; then
                    echo "Current block number: $BLOCK_NUMBER (waiting for >= 5)"
                fi
                
                # Check if block number is >= 5
                if [ "$BLOCK_NUMBER" -ge 5 ]; then
                    echo "Node is ready! Block number: $BLOCK_NUMBER"
                    break
                fi
            fi
        fi
    fi
    
    if [ "$VERBOSE" = true ]; then
        echo "Waiting for node... ($ELAPSED/$TIMEOUT seconds)"
    fi
    
    sleep 2
    ELAPSED=$((ELAPSED + 2))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "Error: Node failed to reach block 5 within $TIMEOUT seconds"
    if [ "$VERBOSE" = true ]; then
        echo "Last response: $RESPONSE"
        echo "Checking node logs:"
        tail -20 /tmp/evmd.log 2>/dev/null || echo "No evmd logs found"
    fi
    exit 1
fi

# Change to the test directory
cd "$TEST_DIR"

# Source the environment file
if [ -f ".env" ]; then
    echo "Sourcing .env file..."
    source .env
else
    echo "Error: .env file not found in $TEST_DIR"
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
forge script script/DeployNFTDescriptor.s.sol:DeployNFTDescriptor \
  --rpc-url "$CUSTOM_RPC" \
  --broadcast \
  --chain-id "$CHAIN_ID"

# Take a rest to ensure the deployment is complete
echo "Waiting for NFTDescriptor deployment to complete..."
sleep 5

# Deploy UniswapV3 with NFTDescriptor library
echo "Deploying UniswapV3..."
forge script script/DeployUniswapV3.s.sol:DeployUniswapV3 \
  --rpc-url "$CUSTOM_RPC" \
  --chain-id "$CHAIN_ID" \
  --broadcast \
  --slow \
  --private-key "$PRIVATE_KEY" \
  --libraries "lib/v3-periphery/contracts/libraries/NFTDescriptor.sol:NFTDescriptor:$LIBRARY_CONTRACT"

echo "foundry-uniswap-v3 tests completed successfully!"