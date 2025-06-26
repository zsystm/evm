// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.0;

interface Vm {
    function broadcast(uint256 privateKey) external;
    function startBroadcast(uint256 privateKey) external;
    function stopBroadcast() external;
    function envUint(string calldata key) external returns (uint256);
    function addr(uint256 privateKey) external returns (address);
    function etch(address who, bytes calldata code) external;
}

address constant VM_ADDRESS = address(uint160(uint256(keccak256('hevm cheat code'))));
