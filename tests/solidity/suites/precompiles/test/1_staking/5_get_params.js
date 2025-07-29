const { expect } = require('chai')
const hre = require('hardhat')
const {
    STAKING_PRECOMPILE_ADDRESS,
    DEFAULT_GAS_LIMIT
} = require('../common')

describe('StakingI – getParams', function () {
    const GAS_LIMIT = DEFAULT_GAS_LIMIT

    let staking, signer

    before(async () => {
        [signer] = await hre.ethers.getSigners()
        staking = await hre.ethers.getContractAt('StakingI', STAKING_PRECOMPILE_ADDRESS)
    })

    it('should retrieve staking module parameters successfully', async function () {
        const params = await staking.getParams({ gasLimit: GAS_LIMIT })
        
        console.log('Staking params raw:', params)
        console.log('Staking params types:', params.map(p => typeof p))
        
        // Verify that all expected parameters are returned
        expect(params).to.have.lengthOf(6)
        
        const [unbondingTime, maxValidators, maxEntries, historicalEntries, bondDenom, minCommissionRate] = params
        
        console.log('Individual params:')
        console.log('  unbondingTime:', unbondingTime, typeof unbondingTime)
        console.log('  maxValidators:', maxValidators, typeof maxValidators)
        console.log('  maxEntries:', maxEntries, typeof maxEntries)
        console.log('  historicalEntries:', historicalEntries, typeof historicalEntries)
        console.log('  bondDenom:', bondDenom, typeof bondDenom)
        console.log('  minCommissionRate:', minCommissionRate, typeof minCommissionRate)
        
        // Verify parameter types and basic validity
        expect(unbondingTime).to.be.a('bigint')
        expect(unbondingTime).to.be.gt(0n) // Should be positive
        
        // These are uint32 so they come back as bigint in ethers v6
        expect(maxValidators).to.be.a('bigint')
        expect(maxValidators).to.be.gt(0n) // Should be positive
        
        expect(maxEntries).to.be.a('bigint')
        expect(maxEntries).to.be.gt(0n) // Should be positive
        
        expect(historicalEntries).to.be.a('bigint')
        expect(historicalEntries).to.be.gte(0n) // Can be 0 or positive
        
        expect(bondDenom).to.be.a('string')
        expect(bondDenom).to.not.be.empty // Should not be empty
        
        // minCommissionRate is a Dec struct which returns as [value, decimals] array
        if (Array.isArray(minCommissionRate)) {
            expect(minCommissionRate).to.have.lengthOf(2)
            expect(minCommissionRate[0]).to.be.a('bigint') // value
            expect(minCommissionRate[1]).to.be.a('bigint') // decimals
        } else {
            expect(minCommissionRate).to.be.an('object')
            expect(minCommissionRate).to.have.property('value')
        }
        
        console.log('✓ All staking parameters retrieved and validated')
        console.log('Parameter values:')
        console.log('  - Unbonding Time:', unbondingTime.toString(), 'nanoseconds')
        console.log('  - Max Validators:', maxValidators.toString())
        console.log('  - Max Entries:', maxEntries.toString())
        console.log('  - Historical Entries:', historicalEntries.toString())
        console.log('  - Bond Denomination:', bondDenom)
        if (Array.isArray(minCommissionRate)) {
            console.log('  - Min Commission Rate:', minCommissionRate[0].toString(), 'with', minCommissionRate[1].toString(), 'decimals')
        } else {
            console.log('  - Min Commission Rate:', minCommissionRate.value ? minCommissionRate.value.toString() : 'N/A')
        }
    })
})
