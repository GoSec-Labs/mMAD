// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MMadToken.sol";

contract SimpleTransactionsScript is Script {
    address constant MMAD_TOKEN = 0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD;
    
    function run() external {
        vm.startBroadcast(vm.envUint("PRIVATE_KEY"));
        
        MMadToken mmadToken = MMadToken(MMAD_TOKEN);
        
        // Transaction 1: Approve tokens (will succeed)
        mmadToken.approve(address(0x1111111111111111111111111111111111111111), 1000 * 1e18);
        
        // Transaction 2: Approve different amount (will succeed)
        mmadToken.approve(address(0x2222222222222222222222222222222222222222), 5000 * 1e18);
        
        // Transaction 3: Update backing ratio (will succeed - we're admin)
        mmadToken.setMinBackingRatio(120);
        
        // Transaction 4: Another approval (will succeed)
        mmadToken.approve(address(0x3333333333333333333333333333333333333333), 10000 * 1e18);
        
        // Transaction 5: Reset backing ratio (will succeed)
        mmadToken.setMinBackingRatio(110);
        
        vm.stopBroadcast();
    }
}
