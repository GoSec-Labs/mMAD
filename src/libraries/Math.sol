// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title Math
 * @dev Safe mathematical operations and utilities
 */
library Math {
    uint256 internal constant PRECISION = 1e18;
    uint256 internal constant PERCENTAGE_SCALE = 1e4; // 10000 = 100%
    
    /**
     * @dev Calculate percentage with precision
     */
    function percentage(uint256 amount, uint256 percent) internal pure returns (uint256) {
        return (amount * percent) / PERCENTAGE_SCALE;
    }
    
    /**
     * @dev Calculate backing ratio
     */
    function calculateBackingRatio(uint256 reserves, uint256 totalSupply) internal pure returns (uint256) {
        if (totalSupply == 0) return PRECISION;
        return (reserves * PRECISION) / totalSupply;
    }
    
    /**
     * @dev Check if reserves meet minimum ratio
     */
    function meetsMinimumRatio(
        uint256 reserves,
        uint256 totalSupply,
        uint256 minRatio
    ) internal pure returns (bool) {
        uint256 currentRatio = calculateBackingRatio(reserves, totalSupply);
        return currentRatio >= minRatio;
    }
    
    /**
     * @dev Calculate maximum mintable amount given reserves and ratio
     */
    function maxMintableAmount(
        uint256 reserves,
        uint256 currentSupply,
        uint256 minRatio
    ) internal pure returns (uint256) {
        uint256 maxSupply = (reserves * PRECISION) / minRatio;
        return maxSupply > currentSupply ? maxSupply - currentSupply : 0;
    }
    
    /**
     * @dev Safe multiplication with overflow check
     */
    function safeMul(uint256 a, uint256 b) internal pure returns (uint256) {
        if (a == 0) return 0;
        uint256 c = a * b;
        require(c / a == b, "Math: multiplication overflow");
        return c;
    }
    
    /**
     * @dev Safe division with zero check
     */
    function safeDiv(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b > 0, "Math: division by zero");
        return a / b;
    }
}