const { expect } = require('chai')
const hre = require('hardhat')

describe('Evidence Precompile', function () {
  const EVIDENCE_ADDRESS = '0x0000000000000000000000000000000000000807'
  const GAS_LIMIT        = 1_000_000

  let evidence, signer
  let evidenceHash
  let consensusAddress

  before(async () => {
    [signer] = await hre.ethers.getSigners()
    evidence = await hre.ethers.getContractAt('IEvidence', EVIDENCE_ADDRESS)

    // use our signer address as the consensusAddress for testing
    consensusAddress = "cons"
  })

  it('should submit evidence and emit SubmitEvidence event', async function () {
    // choose a concrete timestamp as an integer
    const now = Math.floor(Date.now() / 1000)

    const equivocation = {
      height: 1,           // height (int64)
      time: now,         // time (int64)
      power: 1000,        // power (int64)
      consensusAddress // string
    }

    // note: submitEvidence takes only the Equivocation struct
    const tx      = await evidence
        .connect(signer)
        .submitEvidence(signer.address, equivocation, { gasLimit: GAS_LIMIT })
    const receipt = await tx.wait(1)

    // find and validate the event
    const evt = receipt.logs
        .map(log => {
          try { return evidence.interface.parseLog(log) }
          catch { return null }
        })
        .find(e => e && e.name === 'SubmitEvidence')
    expect(evt, 'SubmitEvidence event should be emitted').to.exist
    expect(evt.args.submitter).to.equal(signer.address)

    evidenceHash = evt.args.hash
    expect(evidenceHash).to.not.equal('0x')

    console.log('Submitted evidence hash:', evidenceHash)
  })

  it('evidence() should return the submitted evidence', async function () {
    // callStatic for view
    const result = await evidence.callStatic.evidence(evidenceHash)

    // height and time come back as BigInt
    expect(result.height).to.equal(1n)
    expect(result.time).to.equal(BigInt(Math.floor(Date.now() / 1000)))  // or store `now` in outer scope if precise
    expect(result.power).to.equal(1000n)
    expect(result.consensusAddress).to.equal(consensusAddress)
  })

  it('getAllEvidence should list the submitted evidence', async function () {
    const pageReq = { key: '0x', offset: 0, limit: 10, countTotal: true, reverse: false }

    // callStatic for view
    const [list, pageRes] = await evidence.callStatic.getAllEvidence(pageReq)

    expect(Array.isArray(list)).to.be.true
    expect(list.length).to.be.greaterThan(0)

    // each item is Equivocation: [height,time,power,consensusAddress]
    const found = list.find(item => item.consensusAddress === consensusAddress)
    expect(found, 'Should find our submitted evidence in list').to.exist

    console.log('All evidence:', list)
    console.log('Page response:', pageRes)
  })
})
