const { expect } = require('chai');
const hre = require('hardhat');

function formatUnbondingDelegation(res) {
    const delegatorAddress = res[0];
    const validatorAddress = res[1];
    const rawEntries = res[2]; // This is an array of Result(6)

    const entries = rawEntries.map(entry => {
        const [
            creationHeight,
            completionTime,
            initialBalance,
            balance,
            unbondingId,
            unbondingOnHoldRefCount,
        ] = entry;

        return {
            creationHeight: Number(creationHeight),
            completionTime: Number(completionTime),
            initialBalance: BigInt(initialBalance.toString()),
            balance: BigInt(balance.toString()),
            unbondingId: Number(unbondingId),
            unbondingOnHoldRefCount: Number(unbondingOnHoldRefCount),
        };
    });

    return {
        delegatorAddress,
        validatorAddress,
        entries,
    };
}


// Happy path for undelegate using staking precompile
// This test delegates a small amount and then undelegates it

describe('Staking - undelegate', function () {
    it('should undelegate previously delegated tokens', async function () {
        const valAddr = 'cosmosvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pw4xyrql';
        const amount = hre.ethers.parseEther('0.001');

        const staking = await hre.ethers.getContractAt(
            'StakingI',
            '0x0000000000000000000000000000000000000800'
        );

        const [signer] = await hre.ethers.getSigners();

        // Delegate
        const tx = await staking
            .connect(signer)
            .delegate(signer.address, valAddr, amount);
        const receipt = await tx.wait(2);
        console.log('Delegate tx hash:', receipt.hash, 'gas:', receipt.gasUsed);

        const before = await staking.unbondingDelegation(signer.address, valAddr);
        const beforeUnbondingDelegation = formatUnbondingDelegation(before)
        const numEntriesBefore = beforeUnbondingDelegation.entries.length;

        // Undelegate immediately
        const unTx = await staking
            .connect(signer)
            .undelegate(signer.address, valAddr, amount);
        const unReceipt = await unTx.wait(2);
        console.log('Undelegate tx hash:', unReceipt.hash, 'gas:', unReceipt.gasUsed);

        const result = await staking.unbondingDelegation(signer.address, valAddr);
        const unbondingDelegation = formatUnbondingDelegation(result)
        console.log('Unbonding Delegation:', unbondingDelegation);
        const numEntriesAfter = unbondingDelegation.entries.length;
        expect(numEntriesAfter).to.equal(
            numEntriesBefore + 1,
            'Number of unbonding entries should increase by 1'
        );
        expect(unbondingDelegation.entries[0].balance).to.equal(
            amount,
            'Unbonding entry balance should match undelegated amount'
        );
    });
});