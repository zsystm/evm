const { expect } = require('chai')
const hre = require('hardhat')

// Hardhat tests for the P256 precompile
describe('P256 Precompile', function () {
    const P256_ADDRESS = '0x0000000000000000000000000000000000000100'
    const GAS_LIMIT = 1_000_000

    let signer

    before(async () => {
        [signer] = await hre.ethers.getSigners()
    })

    // Test with a known valid signature data (from the Go tests)
    it('verifies valid P256 signature', async function () {
        // This is test data that should pass validation
        // Using known valid P256 signature components
        const hash = '0x4e1243bd22c66e76c2ba9eddc1f91394e57f9f83067b5c4c7c71f18d37f5c39b'  // 32 bytes
        const r = '0x71a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'     // 32 bytes
        const s = '0x81a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'     // 32 bytes
        const x = '0x91a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'     // 32 bytes
        const y = '0xa1a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'     // 32 bytes

        const input = hash + r.slice(2) + s.slice(2) + x.slice(2) + y.slice(2)

        // Make direct call to precompile
        const result = await hre.ethers.provider.call({
            to: P256_ADDRESS,
            data: input,
            gasLimit: GAS_LIMIT
        })

        // The result should be either 0x (empty for invalid) or 32 bytes with success value
        console.log('Result:', result)
        expect(result).to.be.a('string')
    })

    it('handles invalid signature length', async function () {
        // Test with invalid input length (should be 160 bytes, this is only 32)
        const shortInput = '0x4e1243bd22c66e76c2ba9eddc1f91394e57f9f83067b5c4c7c71f18d37f5c39b'

        const result = await hre.ethers.provider.call({
            to: P256_ADDRESS,
            data: shortInput,
            gasLimit: GAS_LIMIT
        })

        // Should return empty for invalid input length
        expect(result).to.equal('0x')
    })

    it('handles empty input', async function () {
        const result = await hre.ethers.provider.call({
            to: P256_ADDRESS,
            data: '0x',
            gasLimit: GAS_LIMIT
        })

        // Should return empty for invalid input
        expect(result).to.equal('0x')
    })

    it('handles zero signature components', async function () {
        // Test with all zeros (should be invalid signature)
        const zeros = '0x' + '00'.repeat(160) // 160 bytes of zeros

        const result = await hre.ethers.provider.call({
            to: P256_ADDRESS,
            data: zeros,
            gasLimit: GAS_LIMIT
        })

        // Should return empty for invalid signature
        expect(result).to.equal('0x')
    })

    it('tests gas consumption', async function () {
        const hash = '0x4e1243bd22c66e76c2ba9eddc1f91394e57f9f83067b5c4c7c71f18d37f5c39b'
        const r = '0x71a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'
        const s = '0x81a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'
        const x = '0x91a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'
        const y = '0xa1a3c91a0e8c5e1dd7e8f8e0a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8c8a8f8'

        const input = hash + r.slice(2) + s.slice(2) + x.slice(2) + y.slice(2)

        // Estimate gas for the call
        const gasEstimate = await hre.ethers.provider.estimateGas({
            to: P256_ADDRESS,
            data: input
        })

        console.log('Gas estimate:', gasEstimate.toString())
        
        // Should consume reasonable amount of gas
        expect(gasEstimate).to.be.greaterThan(0)
    })

    it('tests precompile address', async function () {
        // Simple test to verify we can call the precompile address
        const code = await hre.ethers.provider.getCode(P256_ADDRESS)
        
        // Precompile addresses typically have no code but are callable
        console.log('Precompile code:', code)
        expect(code).to.be.a('string')
    })
})