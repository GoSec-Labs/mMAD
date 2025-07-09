// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "lib/forge-std/src/Script.sol";
import "lib/forge-std/src/console.sol";

// Import your contracts
import "../src/MMadToken.sol";
import "../src/ZKReserveVerifier.sol";
import "../src/governance/MMadGovernance.sol";
import "../src/governance/Timelock.sol";

// Import helper
import "./DeployHelper.sol";

contract Deploy is Script, DeployHelper {
    // Contract instances
    MMadToken public mmadToken;
    ZKReserveVerifier public zkVerifier;
    MMadGovernance public governance;
    Timelock public timelock;
    
    // Deployment addresses
    address public deployer;
    address public admin;
    address public reserveManager;
    
    function run() external {
        // Load environment variables
        loadConfig();
        
        console.log("=== MMAD Token Deployment Started ===");
        console.log("Network:", getChainName());
        console.log("Deployer:", deployer);
        console.log("Admin:", admin);
        console.log("Reserve Manager:", reserveManager);
        
        vm.startBroadcast(vm.envUint("PRIVATE_KEY"));
        
        // Deploy contracts in correct order
        deployZKVerifier();
        deployTimelock();
        deployGovernance();
        deployMMadToken();
        
        // Configure contracts
        configureContracts();
        
        vm.stopBroadcast();
        
        // Save deployment addresses
        saveDeployment();
        
        // Verify deployment
        verifyDeployment();
        
        console.log("=== Deployment Completed Successfully! ===");
    }
    
    function deployZKVerifier() internal {
        console.log("\n1. Deploying ZK Verifier...");
        
        zkVerifier = new ZKReserveVerifier(admin);
        
        console.log("ZK Verifier deployed at:", address(zkVerifier));
    }
    
    function deployTimelock() internal {
        console.log("\n2. Deploying Timelock...");
        
        // Timelock constructor params
        uint256 minDelay = vm.envUint("TIMELOCK_DELAY");
        address[] memory proposers = new address[](1);
        address[] memory executors = new address[](1);
        
        proposers[0] = admin; // Will be updated to governance later
        executors[0] = admin;
        
        timelock = new Timelock(minDelay, proposers, executors, admin);
        
        console.log("Timelock deployed at:", address(timelock));
        console.log("Min Delay:", minDelay, "seconds");
    }
    
    function deployGovernance() internal {
        console.log("\n3. Deploying Governance...");
        
        // Note: We'll pass a placeholder for governance token, update after MMadToken deployment
        governance = new MMadGovernance(
            address(0), // Will be updated to MMadToken address
            address(timelock),
            admin
        );
        
        console.log("Governance deployed at:", address(governance));
    }
    
    function deployMMadToken() internal {
        console.log("\n4. Deploying MMAD Token...");
        
        mmadToken = new MMadToken(
            admin,
            reserveManager,
            address(zkVerifier)
        );
        
        console.log("MMAD Token deployed at:", address(mmadToken));
        console.log("Token Name:", mmadToken.name());
        console.log("Token Symbol:", mmadToken.symbol());
        console.log("Token Decimals:", mmadToken.decimals());
    }
    
    function configureContracts() internal {
        console.log("\n5. Configuring Contracts...");
        
        // Update governance with correct token address
        // Note: This would require a governance function to update token address
        // For now, we'll just log that this step is needed
        console.log("Manual step needed: Update governance with token address");
        
        // Grant roles
        console.log("Granting PROPOSER_ROLE to governance...");
        timelock.grantRole(timelock.PROPOSER_ROLE(), address(governance));
        
        console.log("Granting EXECUTOR_ROLE to governance...");
        timelock.grantRole(timelock.EXECUTOR_ROLE(), address(governance));
        
        console.log("Configuration completed!");
    }
    
    function saveDeployment() internal {
        console.log("\n6. Saving Deployment Addresses...");
        
        string memory network = getNetworkSuffix();
        
        // Create deployment file content
        string memory deploymentInfo = string(abi.encodePacked(
            "# Deployment on ", getChainName(), "\n",
            "MMAD_TOKEN_", network, "=", vm.toString(address(mmadToken)), "\n",
            "ZK_VERIFIER_", network, "=", vm.toString(address(zkVerifier)), "\n",
            "GOVERNANCE_", network, "=", vm.toString(address(governance)), "\n",
            "TIMELOCK_", network, "=", vm.toString(address(timelock)), "\n"
        ));
        
        // Save to file
        vm.writeFile("deployments/latest.txt", deploymentInfo);
        
        console.log("Deployment addresses saved to deployments/latest.txt");
    }
    
    function verifyDeployment() internal view {
        console.log("\n7. Verifying Deployment...");
        
        require(address(mmadToken) != address(0), "MMAD Token not deployed");
        require(address(zkVerifier) != address(0), "ZK Verifier not deployed");
        require(address(governance) != address(0), "Governance not deployed");
        require(address(timelock) != address(0), "Timelock not deployed");
        
        // Check basic functionality
        require(mmadToken.hasRole(mmadToken.DEFAULT_ADMIN_ROLE(), admin), "Admin role not set");
        require(mmadToken.zkVerifier() == address(zkVerifier), "ZK Verifier not set");
        require(mmadToken.reserveManager() == reserveManager, "Reserve manager not set");
        
        console.log("All contracts deployed and configured correctly!");
    }
}