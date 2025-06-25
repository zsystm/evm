// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.17;

import "../src/utils/Script.sol";
import "../src/harness/StakingHarness.sol";
import {Coin} from "../../../../precompiles/common/Types.sol";

contract StakingContract is Script {
    string constant VALIDATOR = "cosmosvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pw4xyrql";
    uint256 constant AMOUNT = 1e15; // 0.001 atest assuming 18 decimals

    event SharesAndBalance(uint256 shares, string denom, uint256 amount);

    function run() external {
        uint256 pk = vm.envUint("TEST_PRIVATE_KEY");

        vm.startBroadcast(pk);
        StakingHarness harness = new StakingHarness();
        harness.deposit{value: AMOUNT}();
        bool success = harness.delegateFromSelf(VALIDATOR, AMOUNT);
        require(success, "delegateFromSelf failed");

        // Query on-chain state
        (uint256 shares, Coin memory balance) = harness.delegation(address(harness), VALIDATOR);

        // Log the values via an event
        emit SharesAndBalance(shares, balance.denom, balance.amount);
        vm.stopBroadcast();
    }
}
