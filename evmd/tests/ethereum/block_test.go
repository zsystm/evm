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

package ethereum

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBlockchain(t *testing.T) {
	bt := new(testMatcher)

	// Skip tests incompatible with cosmos/evm
	// MPT/Trie tests - cosmos/evm uses IAVL
	bt.skipLoad(`.*TrieTests.*`)
	bt.skipLoad(`.*trie.*`)

	// Uncle/Ommer tests - PoW specific
	bt.skipLoad(`.*Uncle.*`)
	bt.skipLoad(`.*uncle.*`)
	bt.skipLoad(`.*bcForgedTest/bcForkUncle\.json`)

	// Difficulty tests - PoW specific
	bt.skipLoad(`.*[Dd]ifficulty.*`)

	// Withdrawal tests - Beacon chain specific
	bt.skipLoad(`.*[Ww]ithdrawal.*`)

	// Skip chain reorganization tests that rely on PoW
	bt.skipLoad(`.*bcMultiChainTest/ChainAtoChainB_difficultyB.json`)
	bt.skipLoad(`.*bcMultiChainTest/CallContractFromNotBestBlock.json`)
	bt.skipLoad(`.*bcTotalDifficultyTest.*`)
	bt.skipLoad(`.*bcMultiChainTest/lotsOfLeafs.json`)
	bt.skipLoad(`.*bcFrontierToHomestead/blockChainFrontierWithLargerTDvsHomesteadBlockchain.json`)
	bt.skipLoad(`.*bcFrontierToHomestead/blockChainFrontierWithLargerTDvsHomesteadBlockchain2.json`)
	bt.skipLoad(`.*bcArrowGlacierToParis/powToPosBlockRejection.json`)

	// Skip tests with huge gas usage that might OOM
	bt.skipLoad(`.*randomStatetest94.json.*`)
	bt.skipLoad(`.*/stTimeConsuming/.*`)
	bt.skipLoad(`.*OverflowGasRequire.*`)

	// Skip performance tests
	bt.skipLoad(`^GeneralStateTests/VMTests/vmPerformance`)

	// Slow tests
	bt.slow(`.*bcExploitTest/DelegateCallSpam.json`)
	bt.slow(`.*bcExploitTest/ShanghaiLove.json`)
	bt.slow(`.*bcExploitTest/SuicideIssue.json`)
	bt.slow(`.*/bcForkStressTest/`)
	bt.slow(`.*/bcGasPricerTest/RPC_API_Test.json`)
	bt.slow(`.*/bcWalletTest/`)

	// This directory contains no test.
	bt.skipLoad(`.*\.meta/.*`)

	bt.walk(t, blockTestDir, func(t *testing.T, name string, test *BlockTest) {
		execBlockTest(t, bt, test)
	})
}

// TestExecutionSpecBlocktests runs the test fixtures from execution-spec-tests.
func TestExecutionSpecBlocktests(t *testing.T) {
	if !common.FileExist(executionSpecBlockchainTestDir) {
		t.Skipf("directory %s does not exist", executionSpecBlockchainTestDir)
	}
	bt := new(testMatcher)

	bt.walk(t, executionSpecBlockchainTestDir, func(t *testing.T, name string, test *BlockTest) {
		execBlockTest(t, bt, test)
	})
}

func execBlockTest(t *testing.T, bt *testMatcher, test *BlockTest) {
	// Add panic recovery to show filename when test panics
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PANIC in test file %s: %v", test.Filename, r)
		}
	}()

	// Convert test data to JSON for the cosmos adapter
	testJSON, err := json.Marshal(test.json)
	if err != nil {
		t.Fatalf("Failed to marshal test from file %s: %v", test.Filename, err)
	}

	// Run the test using cosmos/evm
	if err := bt.checkFailure(t, test.RunWithEVMD(t, testJSON)); err != nil {
		// If test fails, print filename for debugging
		t.Logf("Test failed in file: %s", test.Filename)
		t.Errorf("cosmos/evm test failed: %v", err)
	}
}
