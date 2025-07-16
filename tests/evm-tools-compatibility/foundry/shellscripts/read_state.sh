#!/usr/bin/env bash
set -euo pipefail

# Usage: ./read_state.sh <CONTRACT_ADDRESS>
# .env에 CUSTOM_RPC, ALICE_ADDRESS 를 정의해 두세요.

source .env
RPC_URL=${CUSTOM_RPC:-http://127.0.0.1:8545}
CONTRACT=$1
ALICE=${ALICE_ADDRESS:-0x0000000000000000000000000000000000000001}

if [ -z "$CONTRACT" ]; then
  echo "Usage: $0 <CONTRACT_ADDRESS>"
  exit 1
fi

# 1) Chain ID
echo "⛓ Chain ID: $(cast chain-id --rpc-url "$RPC_URL")"

# 2) totalSupply()
echo "🔢 totalSupply: $(cast call --rpc-url "$RPC_URL" "$CONTRACT" 'totalSupply()(uint256)')"

# 3) balanceOf(alice)
echo "👤 balanceOf(alice=$ALICE): $(cast call --rpc-url "$RPC_URL" "$CONTRACT" 'balanceOf(address)(uint256)' "$ALICE")"

# 4) name()
echo "📛 name: $(cast call --rpc-url "$RPC_URL" "$CONTRACT" 'name()(string)')"

# 5) symbol()
echo "🔣 symbol: $(cast call --rpc-url "$RPC_URL" "$CONTRACT" 'symbol()(string)')"

# 6) decimals()
echo "🔢 decimals: $(cast call --rpc-url "$RPC_URL" "$CONTRACT" 'decimals()(uint8)')"