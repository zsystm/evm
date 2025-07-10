const { expect } = require('chai')
const hre = require('hardhat')

// Hardhat tests for the Governance precompile

describe('Gov Precompile', function () {
    const GOV_ADDRESS = '0x0000000000000000000000000000000000000805'
    const GAS_LIMIT = 1_000_000
    const COSMOS_ADDR = 'cosmos1cml96vmptgw99syqrrz8az79xer2pcgp95srxm'
    const GOV_MODULE_ADDR = 'cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn'

    let gov, signer

    before(async () => {
        [signer] = await hre.ethers.getSigners()
        gov = await hre.ethers.getContractAt('IGov', GOV_ADDRESS)
    })

    // helper to craft a minimal bank send proposal in proto-json format
    function buildProposal(toCosmos) {
        const msg = {
            '@type': '/cosmos.bank.v1beta1.MsgSend',
            from_address: GOV_MODULE_ADDR,
            to_address: toCosmos,
            amount: [{ denom: 'atest', amount: '1' }],
        }

        const prop = {
            messages: [msg],
            metadata: 'ipfs://CID',
            title: 'test prop',
            summary: 'test prop',
            expedited: false,
        }

        return Buffer.from(JSON.stringify(prop))
    }

    it('submits a proposal and queries it', async function () {
        const jsonProposal = buildProposal(COSMOS_ADDR)
        const deposit = { denom: 'atest', amount: hre.ethers.parseEther('0.1') }

        const tx = await gov
            .connect(signer)
            .submitProposal(signer.address, jsonProposal, [deposit], { gasLimit: GAS_LIMIT })
        const receipt = await tx.wait(2)

        const evt = receipt.logs
            .map(log => { try { return gov.interface.parseLog(log) } catch { return null } })
            .find(e => e && e.name === 'SubmitProposal')
        expect(evt, 'SubmitProposal event must be emitted').to.exist
        expect(evt.args.proposer).to.equal(signer.address)

        const proposalId = evt.args.proposalId
        const proposal = await gov.getProposal(proposalId)
        expect(proposal.id).to.equal(proposalId)
        expect(proposal.proposer).to.equal(signer.address)
    })

    it('deposits and votes on a proposal', async function () {
        const jsonProposal = buildProposal(COSMOS_ADDR)
        const deposit = { denom: 'atest', amount: hre.ethers.parseEther('1') }

        const submitTx = await gov
            .connect(signer)
            .submitProposal(signer.address, jsonProposal, [deposit], { gasLimit: GAS_LIMIT })
        const submitRcpt = await submitTx.wait(2)
        const submitEvt = submitRcpt.logs
            .map(log => { try { return gov.interface.parseLog(log) } catch { return null } })
            .find(e => e && e.name === 'SubmitProposal')
        const propId = submitEvt.args.proposalId

        const depTx = await gov
            .connect(signer)
            .deposit(signer.address, propId, [deposit], { gasLimit: GAS_LIMIT })
        const depRcpt = await depTx.wait(2)
        const depEvt = depRcpt.logs
            .map(log => { try { return gov.interface.parseLog(log) } catch { return null } })
            .find(e => e && e.name === 'Deposit')
        expect(depEvt, 'Deposit event must be emitted').to.exist
        expect(depEvt.args.proposalId).to.equal(propId)

        const voteTx = await gov
            .connect(signer)
            .vote(signer.address, propId, 1, '', { gasLimit: GAS_LIMIT })
        const voteRcpt = await voteTx.wait(2)
        const voteEvt = voteRcpt.logs
            .map(log => { try { return gov.interface.parseLog(log) } catch { return null } })
            .find(e => e && e.name === 'Vote')
        expect(voteEvt, 'Vote event must be emitted').to.exist
        expect(voteEvt.args.option).to.equal(1)

        const vote = await gov.getVote(propId, signer.address)
        expect(vote.proposalId).to.equal(propId)
        expect(vote.options[0].option).to.equal(1)
    })

    it('queries params and constitution', async function () {
        const params = await gov.getParams()
        expect(params.votingPeriod).to.be.a('bigint')
        expect(params.minDeposit.length).to.be.greaterThan(0)

        const constitution = await gov.getConstitution()
        expect(constitution).to.be.a('string')
    })
})
