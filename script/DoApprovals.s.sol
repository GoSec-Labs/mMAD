// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MMadToken.sol";

contract DoApprovalsScript is Script {
    function run() external {
        vm.startBroadcast(vm.envUint("PRIVATE_KEY"));
        
        MMadToken mmadToken = MMadToken(0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD);
        
        // Approve different amounts to different addresses
        mmadToken.approve(address(0xABCD), 1000 * 1e18);
        mmadToken.approve(address(0xEF12), 2000 * 1e18);
        mmadToken.approve(address(0x3456), 3000 * 1e18);
        
        vm.stopBroadcast();
    }
}
