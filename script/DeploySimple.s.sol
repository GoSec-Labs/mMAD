// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/MMadToken.sol";
import "../src/ZKReserveVerifier.sol";

// Import with aliases
import {Groth16Verifier as ReserveVerifier} from "../src/generated/ReserveProofVerifier.sol";
import {Groth16Verifier as ComplianceVerifier} from "../src/generated/ComplianceCheckVerifier.sol";
import {Groth16Verifier as BatchVerifier} from "../src/generated/BatchVerifierVerifier.sol";

contract DeploySimpleScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log(" Deploying mMAD to Sepolia...");
        console.log(" Deployer:", deployer);
        console.log(" Balance:", deployer.balance / 1e18, "ETH");
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy verifiers
        ReserveVerifier reserveVerifier = new ReserveVerifier();
        ComplianceVerifier complianceVerifier = new ComplianceVerifier();
        BatchVerifier batchVerifier = new BatchVerifier();
        
        // Deploy ZK wrapper
        ZKReserveVerifier zkVerifier = new ZKReserveVerifier(
            deployer,
            address(reserveVerifier),
            address(complianceVerifier),
            address(batchVerifier)
        );
        
        // Deploy mMAD token
        MMadToken mmadToken = new MMadToken(
            deployer,
            deployer,
            address(zkVerifier)
        );
        
        vm.stopBroadcast();
        
        // Display results
        console.log("\n DEPLOYMENT COMPLETE!");
        console.log("ReserveVerifier:   ", address(reserveVerifier));
        console.log("ComplianceVerifier:", address(complianceVerifier));
        console.log("BatchVerifier:     ", address(batchVerifier));
        console.log("ZKReserveVerifier: ", address(zkVerifier));
        console.log("MMadToken:         ", address(mmadToken));
        
        // Verify deployment
        console.log("\n Verification:");
        console.log("Token Name:", mmadToken.name());
        console.log("Token Symbol:", mmadToken.symbol());
        console.log("Max Supply:", mmadToken.maxSupply() / 1e18, "MMAD");
    }
}
