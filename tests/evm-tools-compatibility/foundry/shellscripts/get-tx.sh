#!/usr/bin/env bash

# Usage: ./get-tx.sh <TX_HASH> [RPC_URL]
# Example: ./get-tx.sh 0x1234... http://127.0.0.1:8545

source .env
TX_HASH=$1
RPC_URL=${2:-http://127.0.0.1:8545}

if [ -z "$TX_HASH" ]; then
  echo "Usage: $0 <TX_HASH> [RPC_URL]"
  exit 1
fi

# get transaction by hash
read -r -d '' PAYLOAD <<EOF
{
  "jsonrpc":"2.0",
  "method":"eth_getTransactionByHash",
  "params":["$TX_HASH"],
  "id":1
}
EOF

# get transaction by hash
curl -s -X POST \
  -H "Content-Type: application/json" \
  --data "$PAYLOAD" \
  "$RPC_URL" \
  | jq