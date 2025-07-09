// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "../src/interfaces/IZKVerifier.sol";

/**
 * @title MockZKVerifier
 * @dev Mock ZK verifier for testing that always returns true
 */
contract MockZKVerifier is IZKVerifier {
    uint256 private _maxProofAge = 1 hours;
    uint256 private _requiredReserveRatio = 11000; // 110%
    address private _proofValidator;
    
    function verifyReserveProof(
        ProofData calldata,
        ReserveProof calldata
    ) external pure override returns (bool) {
        return true; // Always pass for testing
    }
    
    function verifyComplianceProof(
        ProofData calldata,
        bytes32,
        uint256
    ) external pure override returns (bool) {
        return true; // Always pass for testing
    }
    
    function verifyBatchProofs(
        ProofData[] calldata,
        bytes32[] calldata
    ) external pure override returns (bool) {
        return true; // Always pass for testing
    }
    
    // Configuration functions (simplified for testing)
    function updateVerificationKey(bytes calldata) external override {}
    function setProofValidator(address validator) external override {
        _proofValidator = validator;
    }
    function setMaxProofAge(uint256 maxAge) external override {
        _maxProofAge = maxAge;
    }
    
    // View functions
    function getLastProofTimestamp() external view override returns (uint256) {
        return block.timestamp;
    }
    
    function isProofValid(bytes32) external pure override returns (bool) {
        return true;
    }
    
    function getRequiredReserveRatio() external view override returns (uint256) {
        return _requiredReserveRatio;
    }
}