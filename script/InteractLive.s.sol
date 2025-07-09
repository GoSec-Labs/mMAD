// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/MMadToken.sol";
import "../src/interfaces/IZKVerifier.sol";

contract InteractLiveScript is Script {
    // the deployed contract address
    address constant MMAD_TOKEN = 0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD;
    
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log(" Interacting with live mMAD on Sepolia...");
        console.log(" Deployer:", deployer);
        
        MMadToken mmadToken = MMadToken(MMAD_TOKEN);
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Transaction 1: Check current state (read-only, no gas)
        console.log("\n Current Token State:");
        console.log("   Name:", mmadToken.name());
        console.log("   Symbol:", mmadToken.symbol());
        console.log("   Total Supply:", mmadToken.totalSupply() / 1e18, "MMAD");
        console.log("   Deployer Balance:", mmadToken.balanceOf(deployer) / 1e18, "MMAD");
        
        // Transaction 2: Setup mock reserves (this will create a transaction)
        setupMockReserves(mmadToken);
        
        // Transaction 3: Mint some tokens (this will create a transaction)
        mintTokens(mmadToken, deployer);
        
        // Transaction 4: Make some transfers (this will create transactions)
        makeTransfers(mmadToken, deployer);
        
        // Transaction 5: Test approval system (this will create transactions)
        testApprovals(mmadToken, deployer);
        
        vm.stopBroadcast();
        
        console.log("\n All transactions completed!");
        console.log(" Check Etherscan for transaction history!");
    }
    
    function setupMockReserves(MMadToken mmadToken) internal {
        console.log("\n Transaction 1: Setting up reserves...");
        
        // Create mock proof (will fail with real verifier, but creates transaction)
        IZKVerifier.ProofData memory mockProof = IZKVerifier.ProofData({
            a: [uint256(1), uint256(2)],
            b: [[uint256(3), uint256(4)], [uint256(5), uint256(6)]],
            c: [uint256(7), uint256(8)],
            publicSignals: new uint256[](1)
        });
        mockProof.publicSignals[0] = 1;
        
        try mmadToken.updateReserves(1_000_000 * 1e18, mockProof) {
            console.log(" Reserves updated successfully");
        } catch Error(string memory reason) {
            console.log(" Reserve update failed (expected):", reason);
            console.log(" Transaction still created on-chain!");
        }
    }
    
    function mintTokens(MMadToken mmadToken, address deployer) internal {
        console.log("\n Transaction 2: Attempting to mint tokens...");
        
        try mmadToken.mint(deployer, 1000 * 1e18) {
            console.log(" Minted 1000 MMAD successfully!");
            console.log("   New balance:", mmadToken.balanceOf(deployer) / 1e18, "MMAD");
        } catch Error(string memory reason) {
            console.log(" Minting failed:", reason);
            console.log(" Transaction still created on-chain!");
        }
    }
    
    function makeTransfers(MMadToken mmadToken, address deployer) internal {
        console.log("\n Transaction 3: Testing transfers...");
        
        uint256 balance = mmadToken.balanceOf(deployer);
        if (balance > 0) {
            address testRecipient = address(0x1234567890123456789012345678901234567890);
            uint256 transferAmount = 100 * 1e18;
            
            try mmadToken.transfer(testRecipient, transferAmount) {
                console.log(" Transferred", transferAmount / 1e18, "MMAD");
            } catch Error(string memory reason) {
                console.log(" Transfer failed:", reason);
                console.log(" Transaction still created on-chain!");
            }
        } else {
            // Try transfer with 0 balance (will fail but create transaction)
            try mmadToken.transfer(address(0x1111), 1) {
                console.log(" Unexpected success");
            } catch Error(string memory reason) {
                console.log(" Transfer failed as expected:", reason);
                console.log(" Transaction created on-chain!");
            }
        }
    }
    
    function testApprovals(MMadToken mmadToken, address deployer) internal {
        console.log("\n Transaction 4: Testing approvals...");
        
        address spender = address(0x2222222222222222222222222222222222222222);
        uint256 approvalAmount = 500 * 1e18;
        
        try mmadToken.approve(spender, approvalAmount) {
            console.log(" Approved", approvalAmount / 1e18, "MMAD for spender");
            console.log("   Allowance:", mmadToken.allowance(deployer, spender) / 1e18, "MMAD");
        } catch Error(string memory reason) {
            console.log(" Approval failed:", reason);
        }
    }
}
