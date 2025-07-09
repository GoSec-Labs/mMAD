// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/MMadToken.sol";
import "../src/ZKReserveVerifier.sol";

// Import with aliases to avoid naming conflicts
import {Groth16Verifier as ReserveVerifier} from "../src/generated/ReserveProofVerifier.sol";
import {Groth16Verifier as ComplianceVerifier} from "../src/generated/ComplianceCheckVerifier.sol";
import {Groth16Verifier as BatchVerifier} from "../src/generated/BatchVerifierVerifier.sol";

contract DeployToSepoliaScript is Script {
    // Contract instances
    ReserveVerifier public reserveVerifier;
    ComplianceVerifier public complianceVerifier;
    BatchVerifier public batchVerifier;
    ZKReserveVerifier public zkVerifier;
    MMadToken public mmadToken;
    
    // Deployment addresses
    address public deployer;
    
    function run() external {
        // Get deployer address from private key
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        deployer = vm.addr(deployerPrivateKey);
        
        console.log(" Starting mMAD deployment to Sepolia...");
        console.log(" Deployer address:", deployer);
        console.log(" Deployer balance:", deployer.balance / 1e18, "ETH");
        
        // Start broadcasting transactions
        vm.startBroadcast(deployerPrivateKey);
        
        // Step 1: Deploy ZK Verifiers
        deployZKVerifiers();
        
        // Step 2: Deploy ZK Wrapper
        deployZKWrapper();
        
        // Step 3: Deploy mMAD Token
        deployMMadToken();
        
        // Step 4: Initialize system
        initializeSystem();
        
        vm.stopBroadcast();
        
        // Step 5: Display results
        displayResults();
        
        // Step 6: Save deployment info
        saveDeploymentInfo();
    }
    
    function deployZKVerifiers() internal {
        console.log("\n Step 1: Deploying ZK Verifiers...");
        
        // Deploy Reserve Proof Verifier
        reserveVerifier = new ReserveVerifier();
        console.log(" ReserveVerifier deployed:", address(reserveVerifier));
        
        // Deploy Compliance Check Verifier
        complianceVerifier = new ComplianceVerifier();
        console.log(" ComplianceVerifier deployed:", address(complianceVerifier));
        
        // Deploy Batch Verifier
        batchVerifier = new BatchVerifier();
        console.log(" BatchVerifier deployed:", address(batchVerifier));
    }
    
    function deployZKWrapper() internal {
        console.log("\n Step 2: Deploying ZK Wrapper...");
        
        zkVerifier = new ZKReserveVerifier(
            deployer,                    // admin
            address(reserveVerifier),    // reserve verifier
            address(complianceVerifier), // compliance verifier
            address(batchVerifier)       // batch verifier
        );
        
        console.log(" ZKReserveVerifier deployed:", address(zkVerifier));
    }
    
    function deployMMadToken() internal {
        console.log("\n Step 3: Deploying mMAD Token...");
        
        mmadToken = new MMadToken(
            deployer,              // admin
            deployer,              // reserve manager (can be changed later)
            address(zkVerifier)    // zk verifier
        );
        
        console.log(" MMadToken deployed:", address(mmadToken));
    }
    
    function initializeSystem() internal {
        console.log("\n  Step 4: Initializing system...");
        
        // Verify token properties
        require(
            keccak256(bytes(mmadToken.name())) == keccak256("Moroccan Mad Stablecoin"),
            "Token name mismatch"
        );
        require(
            keccak256(bytes(mmadToken.symbol())) == keccak256("MMAD"),
            "Token symbol mismatch"
        );
        require(mmadToken.decimals() == 18, "Token decimals mismatch");
        
        console.log(" Token properties verified");
        console.log("   Name:", mmadToken.name());
        console.log("   Symbol:", mmadToken.symbol());
        console.log("   Decimals:", mmadToken.decimals());
        console.log("   Max Supply:", mmadToken.maxSupply() / 1e18, "MMAD");
        
        // Verify ZK integration
        require(mmadToken.zkVerifier() == address(zkVerifier), "ZK verifier mismatch");
        console.log(" ZK integration verified");
        
        // Verify access control
        require(mmadToken.hasRole(mmadToken.DEFAULT_ADMIN_ROLE(), deployer), "Admin role not set");
        require(mmadToken.hasRole(mmadToken.MINTER_ROLE(), deployer), "Minter role not set");
        require(mmadToken.hasRole(mmadToken.PAUSER_ROLE(), deployer), "Pauser role not set");
        console.log(" Access control verified");
    }
    
    function displayResults() internal view {
        console.log("\n DEPLOYMENT COMPLETE!");
        console.log("====================================");
        console.log(" Contract Addresses:");
        console.log("   ReserveVerifier:    ", address(reserveVerifier));
        console.log("   ComplianceVerifier: ", address(complianceVerifier));
        console.log("   BatchVerifier:      ", address(batchVerifier));
        console.log("   ZKReserveVerifier:  ", address(zkVerifier));
        console.log("   MMadToken:          ", address(mmadToken));
        console.log("====================================");
        console.log(" Network: Sepolia Testnet");
        console.log(" Deployer:", deployer);
        console.log(" Remaining balance:", deployer.balance / 1e18, "ETH");
        console.log("====================================");
    }
    
    function saveDeploymentInfo() internal {
        string memory deploymentInfo = string.concat(
            "# mMAD Sepolia Deployment\n\n",
            "**Network:** Sepolia Testnet\n",
            "**Deployer:** ", vm.toString(deployer), "\n",
            "**Block:** ", vm.toString(block.number), "\n",
            "**Timestamp:** ", vm.toString(block.timestamp), "\n\n",
            "## Contract Addresses\n\n",
            "- **ReserveVerifier:** ", vm.toString(address(reserveVerifier)), "\n",
            "- **ComplianceVerifier:** ", vm.toString(address(complianceVerifier)), "\n",
            "- **BatchVerifier:** ", vm.toString(address(batchVerifier)), "\n",
            "- **ZKReserveVerifier:** ", vm.toString(address(zkVerifier)), "\n",
            "- **MMadToken:** ", vm.toString(address(mmadToken)), "\n\n",
            "## Etherscan Links\n\n",
            "- [ReserveVerifier](https://sepolia.etherscan.io/address/", vm.toString(address(reserveVerifier)), ")\n",
            "- [ComplianceVerifier](https://sepolia.etherscan.io/address/", vm.toString(address(complianceVerifier)), ")\n",
            "- [BatchVerifier](https://sepolia.etherscan.io/address/", vm.toString(address(batchVerifier)), ")\n",
            "- [ZKReserveVerifier](https://sepolia.etherscan.io/address/", vm.toString(address(zkVerifier)), ")\n",
            "- [MMadToken](https://sepolia.etherscan.io/address/", vm.toString(address(mmadToken)), ")\n"
        );
        
        // Note: In a real script, you'd write this to a file
        console.log("\n Deployment info saved (copy from logs above)");
    }
}
