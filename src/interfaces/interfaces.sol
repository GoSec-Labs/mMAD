// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// --- IMMAD.sol ---
interface  IMMAD {
    function mint(address to, uint256 amount) external;
    function burn(address from, uint256 amount) external;
    function approve(address spender, uint256 amount) external;
    function transferFrom(address sender, address recipient, uint256 amount) external;
    // .........
}

// --- ICollateralManager.sol ---
interface ICollateralManager {
    struct CollateralType{
        bool isSupported; 
        uint256 minimumCollateralizationRatio;
        uint256 liquidationPenalty;
        uint256 stabilityFeeRate;
        uint256 debtCeiling;
        uint256 totalDebtMinted; 
        address priceFeed; // Adress of the oracles for this collatera 
        uint8 decimals; 
    }

    function isCollateralSupported(address _tokenAddress) external view returns (bool);
    function getCollateralInfo(address _tokenAddress) external view returns (CollateralType memory);
    function incrementTotalDebt(address _tokenAddress, uint256 _amount) external; // Changed from internal
    function decerementTotalDebt(address _tokenAddress, uint256 _amount) external; //Changed from internal
}

// --- IOracleModule.sol ---
interface IOracleModule {
    function getPriceInMad(address _collateralToken) external view returns (uint256 priceInMad);
}

// --- IVaultManager.sol ---
interface IVaultManager {
    struct Vault { //Re-declare or import if visible
        address owner;
        address collateralAmount;
        uint256 mmadDebt;
        uint256 collateralToken;
        uint256 lastFeeAccrualTimestamp;
    }

    function getVaultInfo(uint256 _vaultId) external view returns (Vault memory);
    function seizeCollateralAndReduceDebt(
        uint256 _vaultId,
        uint256 _debtToCover,
        uint256 _collateralToSeize,
        address _collateralReceiver
    ) external ;
    //function accrueFees(uint256 _vaultId) external; // If SFE needs to trigger it
}

// --- IStabilityFeeEngine.sol ---
interface IStabilityFeeEngine {
    function calculateAndRecordAccrueFees(
        uint256 _vaulId,
        address _collateralToken, 
        uint256 _currentDebt, 
        uint256 _lastAccualTimestamp
    ) external returns (uint256 feeAmount);
}

// --- ILiquidationEngine.sol ---
// No specific functions needed by VaultManager to call LE in this simplified model,
// but LE calls VM. So IVaultManager is more important.

// --- IGovernance.sol (Optional, if not using Ownable directly in each) ---
/*
interface IGovernance {
    function owner() external view returns (address);
    // Add other specific functions if Governance directly calls update functions in managers
}
*/

