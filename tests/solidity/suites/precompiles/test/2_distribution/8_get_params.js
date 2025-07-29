const { expect } = require('chai')
const hre = require('hardhat')
const {
    DISTRIBUTION_PRECOMPILE_ADDRESS,
    DEFAULT_GAS_LIMIT
} = require('../common')

describe('DistributionI – getParams', function () {
    const GAS_LIMIT = DEFAULT_GAS_LIMIT

    let distribution, signer

    before(async () => {
        [signer] = await hre.ethers.getSigners()
        distribution = await hre.ethers.getContractAt('DistributionI', DISTRIBUTION_PRECOMPILE_ADDRESS)
    })

    it('should retrieve distribution module parameters successfully', async function () {
        const params = await distribution.getParams({ gasLimit: GAS_LIMIT })
        
        console.log('Distribution params raw:', params)
        console.log('Distribution params types:', params.map(p => typeof p))
        
        // Verify that all expected parameters are returned
        expect(params).to.have.lengthOf(2)
        
        const [communityTax, withdrawAddrEnabled] = params
        
        console.log('Individual params:')
        console.log('  communityTax:', communityTax, typeof communityTax)
        console.log('  withdrawAddrEnabled:', withdrawAddrEnabled, typeof withdrawAddrEnabled)
        
        // The communityTax appears to be returning as a Dec struct which is [value, decimals]
        if (Array.isArray(communityTax)) {
            expect(communityTax).to.have.lengthOf(2)
            expect(communityTax[0]).to.be.a('bigint') // value
            expect(communityTax[1]).to.be.a('bigint') // decimals  
        } else {
            expect(communityTax).to.be.an('object')
            expect(communityTax).to.have.property('value')
        }
        
        expect(withdrawAddrEnabled).to.be.a('boolean')
        
        console.log('✓ All distribution parameters retrieved and validated')
        if (Array.isArray(communityTax)) {
            console.log('  - Community Tax:', communityTax[0].toString(), 'with', communityTax[1].toString(), 'decimals')
        } else {
            console.log('  - Community Tax:', communityTax.value ? communityTax.value.toString() : 'N/A')
        }
        console.log('  - Withdraw Address Enabled:', withdrawAddrEnabled)
    })
})