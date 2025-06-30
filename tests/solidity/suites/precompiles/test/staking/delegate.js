const { expect } = require('chai')
const hre = require('hardhat')

describe('Staking', function () {
  it('should stake ATOM to a validator', async function () {
    const valAddr = 'cosmosvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pw4xyrql'
    const stakeAmount = hre.ethers.parseEther('0.001')

    const staking = await hre.ethers.getContractAt(
        'StakingI',
        '0x0000000000000000000000000000000000000800'
    )

    const [signer] = await hre.ethers.getSigners()

    // Query delegation before staking
    const before = await staking.delegation(signer.address, valAddr)
    const initial = before.balance.amount
    console.log('Initial delegation:', initial.toString())

    const tx = await staking
        .connect(signer)
        .delegate(signer.address, valAddr, stakeAmount)
    const receipt = await tx.wait(2)
    console.log('Delegate tx hash:', receipt.hash, 'gas used:', receipt.gasUsed)

    // Query delegation after staking
    const after = await staking.delegation(signer.address, valAddr)
    console.log('Delegated amount:', after.balance.amount.toString())
    expect(after.balance.amount).to.equal(
        initial + stakeAmount,
        'Delegation balance should increase by staked amount'
    )
  })
})