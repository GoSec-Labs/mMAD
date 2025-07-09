// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MMadToken.sol";

contract DoAdminFunctionsScript is Script {
    function run() external {
        vm.startBroadcast(vm.envUint("PRIVATE_KEY"));
        
        MMadToken mmadToken = MMadToken(0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD);
        
        // Change backing ratio
        mmadToken.setMinBackingRatio(125);
        
        // Change it back
        mmadToken.setMinBackingRatio(110);
        
        vm.stopBroadcast();
    }
}
