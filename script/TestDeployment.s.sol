// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/MMadToken.sol";

contract TestDeploymentScript is Script {
    // Update this address after deployment
    address constant MMAD_TOKEN = 0x0000000000000000000000000000000000000000; // Update after deployment
    
    function run() external view {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log(" Testing deployed mMAD contracts...");
        console.log(" Deployer address:", deployer);
        
        MMadToken mmadToken = MMadToken(MMAD_TOKEN);
        
        // Just read contract state (no transactions)
        console.log("\n Contract Information:");
        console.log("   Name:", mmadToken.name());
        console.log("   Symbol:", mmadToken.symbol());
        console.log("   Decimals:", mmadToken.decimals());
        console.log("   Total Supply:", mmadToken.totalSupply() / 1e18, "MMAD");
        console.log("   Max Supply:", mmadToken.maxSupply() / 1e18, "MMAD");
        console.log("   Deployer Balance:", mmadToken.balanceOf(deployer) / 1e18, "MMAD");
        
        // Check access control
        console.log("\n Access Control:");
        console.log("   Is Admin:", mmadToken.hasRole(mmadToken.DEFAULT_ADMIN_ROLE(), deployer));
        console.log("   Is Minter:", mmadToken.hasRole(mmadToken.MINTER_ROLE(), deployer));
        console.log("   Is Pauser:", mmadToken.hasRole(mmadToken.PAUSER_ROLE(), deployer));
        
        // Check reserve info
        (uint256 totalReserves, uint256 requiredReserves, uint256 backingRatio,) = mmadToken.getReserveInfo();
        console.log("\n Reserve Information:");
        console.log("   Total Reserves:", totalReserves / 1e18, "MMAD");
        console.log("   Required Reserves:", requiredReserves / 1e18, "MMAD");
        console.log("   Backing Ratio:", backingRatio, "%");
        
        console.log("\n Contract state verification complete!");
    }
}
