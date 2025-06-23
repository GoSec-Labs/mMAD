// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./IZKVerifier.sol";


/**
 * @title IMMadToken
 * @dev Main interface for MMad stablecoin with ZK integration
 */
interface IMMadToken is IZKVerifier {
    // Minting with ZK proof requirement
    function mintWithProof(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData
    ) external;
    
    // Reserve management
    function updateReserves(
        uint256 newReserveAmount,
        IZKVerifier.ProofData calldata proof
    ) external;
    
    function getReserveInfo() external view returns (
        uint256 totalReserves,
        uint256 requiredReserves,
        uint256 backingRatio,
        uint256 lastProofTimestamp
    );
    
    // Compliance integration
    function transferWithCompliance(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata complianceProof,
        bytes32 userHash
    ) external returns (bool);
    
    // Emergency functions
    function pause() external;
    function unpause() external;
    function emergencyWithdraw(address token, uint256 amount) external;
    
    // Configuration
    function setZKVerifier(address verifier) external;
    function setReserveManager(address manager) external;
    function setMinBackingRatio(uint256 ratio) external;
    
    // View functions
    function zkVerifier() external view returns (address);
    function reserveManager() external view returns (address);
    function minBackingRatio() external view returns (uint256);
    function isPaused() external view returns (bool);
    function maxSupply() external view returns (uint256);
    
    // Events
    event ReservesUpdated(
        uint256 newAmount,
        uint256 backingRatio,
        uint256 timestamp
    );
    event ZKVerifierUpdated(address oldVerifier, address newVerifier);
    event ReserveManagerUpdated(address oldManager, address newManager);
    event MinBackingRatioUpdated(uint256 oldRatio, uint256 newRatio);
    event ComplianceTransfer(
        address indexed from,
        address indexed to,
        uint256 amount,
        bytes32 userHash
    );
}
