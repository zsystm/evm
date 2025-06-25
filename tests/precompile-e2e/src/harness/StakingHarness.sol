// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../../../../precompiles/staking/StakingI.sol";

contract StakingHarness {
    event Done(string step);

    /// @dev allow funding the contract
    function deposit() external payable {
        emit Done("deposit");
    }

    /// @dev delegate tokens on behalf of this contract
    function delegateFromSelf(string memory validator, uint256 amount) external returns (bool) {
        bool ok = STAKING_CONTRACT.delegate(address(this), validator, amount);
        require(ok, "delegate failed");
        return ok;
    }

    function delegation(address delegator, string calldata validator) external view returns (uint256 shares, Coin memory balance) {
        return STAKING_CONTRACT.delegation(delegator, validator);
    }
}
