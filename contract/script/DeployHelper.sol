// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "lib/forge-std/src/Script.sol";

contract DeployHelper is Script {
    // Network configuration
    uint256 public chainId;
    
    function loadConfig() internal {
        chainId = block.chainid;
        
        // Load addresses based on network
        if (chainId == 31337) {
            // Anvil local network
            deployer = vm.envAddress("DEPLOYER_ADDRESS");
            admin = vm.envAddress("ADMIN_ADDRESS");
            reserveManager = vm.envAddress("RESERVE_MANAGER_ADDRESS");
        } else if (chainId == 97) {
            // BSC Testnet
            deployer = vm.addr(vm.envUint("BSC_TESTNET_PRIVATE_KEY"));
            admin = deployer; // Use deployer as admin for testnet
            reserveManager = deployer; // Use deployer as reserve manager for testnet
        } else {
            revert("Unsupported network");
        }
    }
    
    function getChainName() internal view returns (string memory) {
        if (chainId == 31337) return "Anvil Local";
        if (chainId == 97) return "BSC Testnet";
        if (chainId == 56) return "BSC Mainnet";
        return "Unknown Network";
    }
    
    function getNetworkSuffix() internal view returns (string memory) {
        if (chainId == 31337) return "ANVIL";
        if (chainId == 97) return "TESTNET";
        if (chainId == 56) return "MAINNET";
        return "UNKNOWN";
    }
    
    function getGasPrice() internal view returns (uint256) {
        if (chainId == 31337) return vm.envUint("ANVIL_GAS_PRICE");
        if (chainId == 97) return vm.envUint("BSC_GAS_PRICE");
        return 20 gwei; // Default
    }
    
    function getGasLimit() internal view returns (uint256) {
        if (chainId == 31337) return vm.envUint("ANVIL_GAS_LIMIT");
        if (chainId == 97) return vm.envUint("BSC_GAS_LIMIT");
        return 8000000; // Default
    }
    
    // Deployment state variables (accessible by inheriting contracts)
    address internal deployer;
    address internal admin;
    address internal reserveManager;
}