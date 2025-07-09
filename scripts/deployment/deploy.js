const { ethers } = require("hardhat");

async function main() {
    console.log("🚀 Deploying mMAD contracts...");
    
    const [deployer] = await ethers.getSigners();
    console.log("📝 Deploying with account:", deployer.address);
    
    // 1. Deploy ZKReserveVerifier
    console.log("📦 Deploying ZKReserveVerifier...");
    const ZKReserveVerifier = await ethers.getContractFactory("ZKReserveVerifier");
    const zkVerifier = await ZKReserveVerifier.deploy(deployer.address);
    await zkVerifier.waitForDeployment();
    console.log("✅ ZKReserveVerifier deployed to:", await zkVerifier.getAddress());
    
    // 2. Deploy MMadToken
    console.log("📦 Deploying MMadToken...");
    const MMadToken = await ethers.getContractFactory("MMadToken");
    const mmadToken = await MMadToken.deploy(
        deployer.address,      // admin
        deployer.address,      // reserve manager
        await zkVerifier.getAddress()  // zk verifier
    );
    await mmadToken.waitForDeployment();
    console.log("✅ MMadToken deployed to:", await mmadToken.getAddress());
    
    // 3. Deploy Governance (optional)
    console.log("📦 Deploying Governance...");
    const MMadGovernance = await ethers.getContractFactory("MMadGovernance");
    const governance = await MMadGovernance.deploy(
        await mmadToken.getAddress(),  // governance token
        deployer.address,              // timelock (simplified)
        deployer.address               // admin
    );
    await governance.waitForDeployment();
    console.log("✅ Governance deployed to:", await governance.getAddress());
    
    // Save deployment info
    const deploymentInfo = {
        zkVerifier: await zkVerifier.getAddress(),
        mmadToken: await mmadToken.getAddress(),
        governance: await governance.getAddress(),
        deployer: deployer.address,
        network: network.name,
        blockNumber: await ethers.provider.getBlockNumber(),
        timestamp: new Date().toISOString()
    };
    
    console.log("\n📋 Deployment Summary:");
    console.log(JSON.stringify(deploymentInfo, null, 2));
    
    // Save to file
    require("fs").writeFileSync(
        "scripts/deployment/latest-deployment.json", 
        JSON.stringify(deploymentInfo, null, 2)
    );
    
    console.log("\n🎉 Deployment complete!");
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
