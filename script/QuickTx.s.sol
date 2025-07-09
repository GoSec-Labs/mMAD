// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MMadToken.sol";

contract QuickTxScript is Script {
    function run() external {
        vm.broadcast(vm.envUint("PRIVATE_KEY"));
        MMadToken(0xC5a1a52AC838EF30db179c25F3D4a9E750F42ABD).approve(address(0x999), 1000 * 1e18);
    }
}
