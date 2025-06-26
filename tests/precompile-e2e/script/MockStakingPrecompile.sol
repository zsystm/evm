// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.17;

// adjust this path to wherever your real interface lives
import "./StakingI.sol";

contract MockStakingPrecompile is StakingI {
    // --- Non-view methods (no `pure`/`view`) ---

    function createValidator(
        Description calldata,
        CommissionRates calldata,
        uint256,
        address validatorAddress,
        string calldata,
        uint256 value
    ) external override returns (bool) {
        emit CreateValidator(validatorAddress, value);
        return true;
    }

    function editValidator(
        Description calldata,
        address validatorAddress,
        int256 commissionRate,
        int256 minSelfDelegation
    ) external override returns (bool) {
        emit EditValidator(validatorAddress, commissionRate, minSelfDelegation);
        return true;
    }

    function delegate(
        address delegatorAddress,
        string memory validatorAddress,
        uint256 amount
    ) external override returns (bool) {
        // reuse delegatorAddress as the indexed validatorAddress in the event
        emit Delegate(delegatorAddress, delegatorAddress, amount, amount);
        return true;
    }

    function undelegate(
        address delegatorAddress,
        string calldata,
        uint256 amount
    ) external override returns (int64) {
        uint256 ts = block.timestamp;
        emit Unbond(delegatorAddress, delegatorAddress, amount, ts);
        return int64(int256(ts));
    }

    function redelegate(
        address delegatorAddress,
        string calldata,
        string calldata,
        uint256 amount
    ) external override returns (int64) {
        uint256 ts = block.timestamp;
        emit Redelegate(delegatorAddress, delegatorAddress, delegatorAddress, amount, ts);
        return int64(int256(ts));
    }

    function cancelUnbondingDelegation(
        address delegatorAddress,
        string calldata,
        uint256 amount,
        uint256 creationHeight
    ) external override returns (bool) {
        emit CancelUnbondingDelegation(delegatorAddress, delegatorAddress, amount, creationHeight);
        return true;
    }

    // --- View methods (must be marked `view`) ---

    function delegation(
        address,
        string calldata
    ) external view override returns (uint256 shares, Coin memory balance) {
        balance = Coin({ denom: "", amount: 0 });
        return (0, balance);
    }

    function unbondingDelegation(
        address,
        string calldata
    ) external view override returns (UnbondingDelegationOutput memory output) {
        return output;
    }

    function validator(
        address /*validatorAddress*/
    ) external view override returns (Validator memory output) {
        return output;
    }

    function validators(
        string calldata,
        PageRequest calldata
    ) external view override returns (Validator[] memory vs, PageResponse memory pr) {
        return (vs, pr);
    }

    function redelegation(
        address,
        string calldata,
        string calldata
    ) external view override returns (RedelegationOutput memory output) {
        return output;
    }

    function redelegations(
        address,
        string calldata,
        string calldata,
        PageRequest calldata
    ) external view override returns (RedelegationResponse[] memory rs, PageResponse memory pr) {
        return (rs, pr);
    }
}