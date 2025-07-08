// SPDX-License-Identifier: LGPL-v3
pragma solidity >=0.8.17;

import "./../../ics20/ICS20I.sol";
import "./../../common/Types.sol";

contract InterchainSender {
    int64 public counter;

    /// @dev transfer a given amount of tokens. Returns the IBC packet sequence of the IBC transaction.
    /// @dev This emits a IBCTransfer event.
    /// @param sourcePort The source port of the IBC transfer.
    /// @param sourceChannel The source channel of the IBC transfer.
    /// @param denom The denomination of the tokens to transfer.
    /// @param receiver The receiver address on the receiving chain.
    /// @param timeoutHeight The timeout height for the IBC packet.
    /// @param timeoutTimestamp The timeout timestamp of the IBC packet.
    /// @param memo The IBC transaction memo.
    /// @param amount The amount of tokens to transfer to another chain.
    /// @return nextSequence The IBC transaction sequence number.
    function testTransferUserFunds(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        Height memory timeoutHeight,
        uint64 timeoutTimestamp,
        string memory memo
    ) public returns (uint64 nextSequence) {
        return
            ICS20_CONTRACT.transfer(
                sourcePort,
                sourceChannel,
                denom,
                amount,
                msg.sender,
                receiver,
                timeoutHeight,
                timeoutTimestamp,
                memo
            );
    }

    function testTransferContractFunds(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        Height memory timeoutHeight,
        uint64 timeoutTimestamp,
        string memory memo
    ) public returns (uint64 nextSequence) {
        return
            ICS20_CONTRACT.transfer(
                sourcePort,
                sourceChannel,
                denom,
                amount,
                address(this),
                receiver,
                timeoutHeight,
                timeoutTimestamp,
                memo
            );
    }


    function testMultiTransferWithInternalTransfer(
        address payable _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        bool _before,
        bool _between,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = _source.call{value: 15000000000000}("");
            require(sent, "Failed to send Ether to delegator");
        }
        Height memory timeoutHeight = Height(100, 100);
        ICS20_CONTRACT.transfer(
            sourcePort,
            sourceChannel,
            denom,
            amount / 2,
            _source,
            receiver,
            timeoutHeight,
            0,
            ""
        );
        if (_between) {
            counter++;
            (bool sent, ) = _source.call{value: 15000000000000}("");
            require(sent, "Failed to send Ether to delegator");
        }
        ICS20_CONTRACT.transfer(
            sourcePort,
            sourceChannel,
            denom,
            amount / 2,
            _source,
            receiver,
            timeoutHeight,
            0,
            ""
        );
        if (_after) {
            counter++;
            (bool sent, ) = _source.call{value: 15000000000000}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }


 function testTransferFundsWithTransferToOtherAcc(
        address payable _otherAcc,
        address _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = _otherAcc.call{value: 15000000000000}("");
            require(sent, "Failed to send Ether to delegator");
        }
        Height memory timeoutHeight = Height(100, 100);
        ICS20_CONTRACT.transfer(
            sourcePort,
            sourceChannel,
            denom,
            amount,
            _source,
            receiver,
            timeoutHeight,
            0,
            ""
        );
        if (_after) {
            counter++;
            (bool sent, ) = _otherAcc.call{value: 15000000000000}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }

    // QUERIES
    function testDenoms(
        PageRequest calldata pageRequest
    )
        public
        view
        returns (
            Denom[] memory denoms,
            PageResponse memory pageResponse
        )
    {
        return ICS20_CONTRACT.denoms(pageRequest);
    }

    function testDenom(
        string memory hash
    ) public view returns (Denom memory denom) {
        return ICS20_CONTRACT.denom(hash);
    }

    function testDenomHash(
        string memory trace
    ) public view returns (string memory hash) {
        return ICS20_CONTRACT.denomHash(trace);
    }

}
