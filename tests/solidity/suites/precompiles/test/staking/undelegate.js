const { expect } = require('chai')
const hre = require('hardhat')

function formatUnbondingDelegation(res) {
    const delegatorAddress = res[0]
    const validatorAddress = res[1]
    const rawEntries = res[2] // array of Result(6)

    const entries = rawEntries.map(entry => {
        const [
            creationHeight,
            completionTime,
            initialBalance,
            balance,
            unbondingId,
            unbondingOnHoldRefCount,
        ] = entry

        return {
            creationHeight:          Number(creationHeight),
            completionTime:          Number(completionTime),
            initialBalance:          BigInt(initialBalance.toString()),
            balance:                 BigInt(balance.toString()),
            unbondingId:             Number(unbondingId),
            unbondingOnHoldRefCount: Number(unbondingOnHoldRefCount),
        }
    })

    return {
        delegatorAddress,
        validatorAddress,
        entries,
    }
}

describe('Staking – delegate and undelegate with event assertions', function () {
    const STAKING_ADDRESS = '0x0000000000000000000000000000000000000800'

    let staking, bech32, signer

    before(async () => {
        [signer] = await hre.ethers.getSigners()

        // Instantiate precompile contracts
        staking = await hre.ethers.getContractAt('StakingI', STAKING_ADDRESS)
    })

    it('should delegate then undelegate and emit correct events', async function () {
        const valBech32 = 'cosmosvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pw4xyrql'
        const amount    = hre.ethers.parseEther('0.001')

        // DELEGATE
        const delegateTx      = await staking.connect(signer).delegate(signer.address, valBech32, amount)
        const delegateReceipt = await delegateTx.wait(2)
        console.log('Delegate tx hash:', delegateReceipt.hash, 'gas used:', delegateReceipt.gasUsed.toString())

        const hexValAddr = '0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E'

        // Parse and assert the Delegate event
        const delegateEvt = delegateReceipt.logs
            .map(log => {
                try { return staking.interface.parseLog(log) }
                catch { return null }
            })
            .find(evt => evt && evt.name === 'Delegate')
        expect(delegateEvt, 'Delegate event should be emitted').to.exist
        expect(delegateEvt.args.delegatorAddress).to.equal(signer.address)
        expect(delegateEvt.args.validatorAddress).to.equal(hexValAddr)
        expect(delegateEvt.args.amount).to.equal(amount)

        // COUNT UNBONDING ENTRIES BEFORE
        const beforeRaw        = await staking.unbondingDelegation(signer.address, valBech32)
        const beforeUnbonding  = formatUnbondingDelegation(beforeRaw)
        const entriesBefore    = beforeUnbonding.entries.length

        // UNDELEGATE
        const undelegateTx      = await staking.connect(signer).undelegate(signer.address, valBech32, amount)
        const undelegateReceipt = await undelegateTx.wait(2)
        console.log('Undelegate tx hash:', undelegateReceipt.hash, 'gas used:', undelegateReceipt.gasUsed.toString())

        // Parse and assert the Unbond event
        const unbondEvt = undelegateReceipt.logs
            .map(log => {
                try { return staking.interface.parseLog(log) }
                catch { return null }
            })
            .find(evt => evt && evt.name === 'Unbond')
        expect(unbondEvt, 'Unbond event should be emitted').to.exist
        expect(unbondEvt.args.delegatorAddress).to.equal(signer.address)
        expect(unbondEvt.args.validatorAddress).to.equal(hexValAddr)
        expect(unbondEvt.args.amount).to.equal(amount)

        // Assert that completionTime is a positive BigInt
        const completionTime = BigInt(unbondEvt.args.completionTime.toString())
        expect(completionTime > 0n, 'completionTime should be positive').to.be.true

        // COUNT UNBONDING ENTRIES AFTER
        const afterRaw       = await staking.unbondingDelegation(signer.address, valBech32)
        const afterUnbonding = formatUnbondingDelegation(afterRaw)
        const entriesAfter   = afterUnbonding.entries.length

        expect(entriesAfter).to.equal(
            entriesBefore + 1,
            'Number of unbonding entries should increase by 1'
        )
        expect(afterUnbonding.entries[0].balance).to.equal(
            BigInt(amount.toString()),
            'Unbonding entry balance should match undelegated amount'
        )
    })
})
