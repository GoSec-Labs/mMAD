// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MMadToken.sol";

contract GenerateActivityScript is Script {
    address constant MMAD_TOKEN = 0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD;
    
    function run() external {
        vm.startBroadcast(vm.envUint("PRIVATE_KEY"));
        
        MMadToken mmadToken = MMadToken(MMAD_TOKEN);
        
        // Generate 5 different transactions to show activity
        
        // Transaction 1: Approve some tokens
        mmadToken.approve(address(0x1111), 1000 * 1e18);
        
        // Transaction 2: Try to mint (will fail but shows transaction)
        try mmadToken.mint(address(0x2222), 100 * 1e18) {} catch {}
        
        // Transaction 3: Update backing ratio
        try mmadToken.setMinBackingRatio(120) {} catch {}
        
        // Transaction 4: Another approval
        mmadToken.approve(address(0x3333), 2000 * 1e18);
        
        // Transaction 5: Try pause (will work since we're admin)
        try mmadToken.pause() {} catch {}
        
        vm.stopBroadcast();
    }
}
