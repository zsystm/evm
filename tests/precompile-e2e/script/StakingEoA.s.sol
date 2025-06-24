// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.17;

import "../src/utils/Script.sol";
import "../../precompiles/staking/StakingI.sol";

contract StakingEOA is Script {
    string constant VALIDATOR = "cosmosvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pw4xyrql";
    uint256 constant AMOUNT = 1e15; // 0.001 atest assuming 18 decimals

    function run() external {
        uint256 pk = vm.envUint("TEST_PRIVATE_KEY");
        address delegator = vm.addr(pk);

        vm.startBroadcast(pk);
        STAKING_CONTRACT.delegate(delegator, VALIDATOR, AMOUNT);
        vm.stopBroadcast();

        (, Coin memory balance) = STAKING_CONTRACT.delegation(delegator, VALIDATOR);
        require(balance.amount == AMOUNT, "incorrect amount");
    }
}
