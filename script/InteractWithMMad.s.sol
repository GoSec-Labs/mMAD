// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/MMadToken.sol";
import "../src/interfaces/IZKVerifier.sol";

contract InteractWithMMadScript is Script {
    // addresses after deployment
    address constant MMAD_TOKEN = 0x0000000000000000000000000000000000000000; 
    
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log(" Interacting with deployed mMAD...");
        console.log(" User address:", deployer);
        
        MMadToken mmadToken = MMadToken(MMAD_TOKEN);
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Test 1: Check token info
        console.log("\n Token Information:");
        console.log("   Name:", mmadToken.name());
        console.log("   Symbol:", mmadToken.symbol());
        console.log("   Total Supply:", mmadToken.totalSupply() / 1e18, "MMAD");
        console.log("   Max Supply:", mmadToken.maxSupply() / 1e18, "MMAD");
        
        // Test 2: Setup reserves (with mock proof)
        setupReserves(mmadToken, deployer);
        
        // Test 3: Mint some tokens
        mintTokens(mmadToken, deployer);
        
        // Test 4: Transfer tokens
        transferTokens(mmadToken, deployer);
        
        vm.stopBroadcast();
        
        console.log("\n All interactions completed successfully!");
    }
    
    function setupReserves(MMadToken mmadToken, address deployer) internal {
        console.log("\n Setting up reserves...");
        
        // Create a mock proof for testing
        IZKVerifier.ProofData memory mockProof = IZKVerifier.ProofData({
            a: [uint256(1), uint256(2)],
            b: [[uint256(3), uint256(4)], [uint256(5), uint256(6)]],
            c: [uint256(7), uint256(8)],
            publicSignals: new uint256[](1)
        });
        mockProof.publicSignals[0] = 1; // Valid proof result
        
        try mmadToken.updateReserves(1_000_000 * 1e18, mockProof) {
            console.log(" Reserves updated: 1,000,000 MMAD backing");
            
            (uint256 totalReserves,,,) = mmadToken.getReserveInfo();
            console.log("   Total Reserves:", totalReserves / 1e18, "MMAD");
        } catch Error(string memory reason) {
            console.log(" Reserve update failed:", reason);
            console.log(" This is expected with real ZK verifier - need valid proof");
        }
    }
    
    function mintTokens(MMadToken mmadToken, address deployer) internal {
        console.log("\n Minting tokens...");
        
        // Use deployer address for minting (they have MINTER_ROLE)
        address mintTo = deployer;
        
        try mmadToken.mint(mintTo, 1000 * 1e18) {
            console.log(" Minted 1000 MMAD to:", mintTo);
            console.log("   Balance:", mmadToken.balanceOf(mintTo) / 1e18, "MMAD");
            console.log("   Total Supply:", mmadToken.totalSupply() / 1e18, "MMAD");
        } catch Error(string memory reason) {
            console.log(" Minting failed:", reason);
            console.log(" Likely needs reserves setup first");
        }
    }
    
    function transferTokens(MMadToken mmadToken, address deployer) internal {
        console.log("\n Testing transfers...");
        
        uint256 balance = mmadToken.balanceOf(deployer);
        console.log("   Current balance:", balance / 1e18, "MMAD");
        
        if (balance > 0) {
            // Use a different test address for recipient
            address recipient = address(0x1234567890123456789012345678901234567890);
            uint256 transferAmount = balance / 10; // Transfer 10% of balance
            
            if (transferAmount > 0) {
                try mmadToken.transfer(recipient, transferAmount) {
                    console.log(" Transferred", transferAmount / 1e18, "MMAD to:", recipient);
                    console.log("   Deployer new balance:", mmadToken.balanceOf(deployer) / 1e18, "MMAD");
                    console.log("   Recipient balance:", mmadToken.balanceOf(recipient) / 1e18, "MMAD");
                } catch Error(string memory reason) {
                    console.log(" Transfer failed:", reason);
                }
            } else {
                console.log(" Transfer amount too small");
            }
        } else {
            console.log(" No tokens to transfer - mint some first");
        }
    }
}