const { ethers } = require("hardhat");

let deployedAddr;

async function deploy(){
    const [deployer, s1, s2, s3, s4, s5] = await ethers.getSigners();
    console.log("Deployer:", deployer.address);
  
    // 1. contract 배포
    const Contract = await ethers.getContractFactory("TokenExample");
    const exampleToken=await Contract.deploy();

    await exampleToken.waitForDeployment();
    const addr = await exampleToken.getAddress();
    console.log('Token deployed to:'+addr)
    deployedAddr = addr;


}

async function main(){
    await deploy();
}

main();