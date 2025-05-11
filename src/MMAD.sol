// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "lib/openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";
import "lib/openzeppelin-contracts/contracts/access/Ownable.sol"; // Or a custom Governance contract

/**
 * @title MMAD Token
 * @author GoSec-Labs/mMAD
 * @notice This is the ERC20 token contract for mMAD, the Moroccan Dirham-pegged stablecoin.
 * Its supply is controlled by the VaultManager contract.
 */
contract MMAD is ERC20, Ownable { // Replace Ownable with your Governance contract if preferred
    address public vaultManager;

    event VaultManagerSet(address indexed newVaultManager);

    /**
     * @notice Constructor to initialize the mMAD token.
     * @param _initialOwner The initial owner of the contract (e.g., deployer or Governance contract).
     */
    constructor(address _initialOwner) ERC20("Moroccan Dirham Test (mMAD)", "mMADt") Ownable(_initialOwner) {
        // mMADt for testnet, can be mMAD on mainnet
    }

    /**
     * @notice Sets the address of the VaultManager contract.
     * Only the owner (Governance) can call this function.
     * @param _vaultManager The address of the VaultManager contract.
     */
    function setVaultManager(address _vaultManager) external onlyOwner {
        require(_vaultManager != address(0), "MMAD: VaultManager address cannot be zero");
        vaultManager = _vaultManager;
        emit VaultManagerSet(_vaultManager);
    }

    /**
     * @notice Mints new mMAD tokens.
     * Only callable by the VaultManager contract.
     * @param to The address to mint tokens to.
     * @param amount The amount of tokens to mint.
     */
    function mint(address to, uint256 amount) external {
        require(msg.sender == vaultManager, "MMAD: Caller is not the VaultManager");
        _mint(to, amount);
    }

    /**
     * @notice Burns mMAD tokens.
     * Only callable by the VaultManager contract.
     * @param from The address to burn tokens from.
     * @param amount The amount of tokens to burn.
     */
    function burn(address from, uint256 amount) external {
        require(msg.sender == vaultManager, "MMAD: Caller is not the VaultManager");
        _burn(from, amount);
    }




    //Need more calls,functions 
}
