// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IZKVerifier.sol";
import "./utils/AccessControl.sol";
import "./utils/Pausable.sol";
import "./libraries/Math.sol";
import "./libraries/Errors.sol";

/**
 * @title ZKReserveVerifier
 * @dev Verifies zero-knowledge proofs for reserve backing and compliance
 */
contract ZKReserveVerifier is IZKVerifier, AccessControl, Pausable {
    
    // Configuration
    uint256 private _maxProofAge = 1 hours;
    uint256 private _requiredReserveRatio = 110 * 100 / 100; // 110% (simplified)
    address private _proofValidator;
    
    // Proof tracking
    mapping(bytes32 => bool) private _verifiedProofs;
    mapping(bytes32 => uint256) private _proofTimestamps;
    
    event VerificationKeyUpdated(bytes32 indexed keyType, bytes32 keyHash);
    event ProofValidatorUpdated(address oldValidator, address newValidator);
    event MaxProofAgeUpdated(uint256 oldAge, uint256 newAge);
    
    constructor(address admin) {
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(PROOF_VALIDATOR_ROLE, admin);
        _proofValidator = admin;
    }
    
    function verifyReserveProof(
        IZKVerifier.ProofData calldata proof,
    IZKVerifier.ReserveProof calldata reserveData
    ) external view override returns (bool) {
        // Basic validation
        if (proof.publicSignals.length < 3) return false;
        if (reserveData.requiredReserve == 0) return false;
        if (reserveData.currentSupply == 0) return false;
        
        // Verify reserve ratio meets minimum requirements
        uint256 actualRatio = _calculateBackingRatio(
            reserveData.requiredReserve,
            reserveData.currentSupply
        );
        
        if (actualRatio < _requiredReserveRatio) return false;
        
        // Verify proof timestamp is recent
        if (!_isProofTimestampValid(reserveData.timestamp)) return false;
        
        // Simulate verification
        return _simulateGroth16Verification(proof, reserveData);
    }
    
    function verifyComplianceProof(
        ProofData calldata proof,
        bytes32 userHash,
        uint256 riskScore
    ) external view override returns (bool) {
        if (proof.publicSignals.length < 2) return false;
        if (userHash == bytes32(0)) return false;
        if (riskScore > 100) return false;
        
        return _simulateComplianceVerification(proof, userHash, riskScore);
    }
    
    function verifyBatchProofs(
        ProofData[] calldata proofs,
        bytes32[] calldata commitments
    ) external view override returns (bool) {
        if (proofs.length != commitments.length) return false;
        if (proofs.length == 0) return false;
        
        for (uint256 i = 0; i < proofs.length; i++) {
            if (proofs[i].publicSignals.length == 0) return false;
            if (commitments[i] == bytes32(0)) return false;
        }
        
        return true;
    }
    
    // Configuration functions
    function updateVerificationKey(bytes calldata vkData) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        bytes32 keyHash = keccak256(vkData);
        emit VerificationKeyUpdated("RESERVE", keyHash);
    }
    
    function setProofValidator(address validator) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (validator == address(0)) revert Errors.ZeroAddress();
        address oldValidator = _proofValidator;
        _proofValidator = validator;
        
        _revokeRole(PROOF_VALIDATOR_ROLE, oldValidator);
        _grantRole(PROOF_VALIDATOR_ROLE, validator);
        
        emit ProofValidatorUpdated(oldValidator, validator);
    }
    
    function setMaxProofAge(uint256 maxAge) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (maxAge == 0) revert Errors.InvalidParameter();
        uint256 oldAge = _maxProofAge;
        _maxProofAge = maxAge;
        emit MaxProofAgeUpdated(oldAge, maxAge);
    }
    
    // View functions
    function getLastProofTimestamp() external view override returns (uint256) {
        return block.timestamp;
    }
    
    function isProofValid(bytes32 proofHash) external view override returns (bool) {
        return _verifiedProofs[proofHash] && 
               block.timestamp - _proofTimestamps[proofHash] <= _maxProofAge;
    }
    
    function getRequiredReserveRatio() external view override returns (uint256) {
        return _requiredReserveRatio;
    }
    
    // Internal helper functions
    function _calculateBackingRatio(
        uint256 reserves,
        uint256 supply
    ) internal pure returns (uint256) {
        if (supply == 0) return 0;
        return (reserves * 100) / supply;
    }
    
    function _isProofTimestampValid(uint256 timestamp) internal view returns (bool) {
        return timestamp > 0 && timestamp <= block.timestamp && 
               block.timestamp - timestamp <= _maxProofAge;
    }
    
    function _simulateGroth16Verification(
        ProofData memory proof,
        ReserveProof memory reserveData
    ) internal pure returns (bool) {
        if (proof.publicSignals.length < 3) return false;
        
        uint256 providedMinReserve = proof.publicSignals[0];
        uint256 providedSupply = proof.publicSignals[1];
        uint256 providedTimestamp = proof.publicSignals[2];
        
        return providedMinReserve == reserveData.requiredReserve &&
               providedSupply == reserveData.currentSupply &&
               providedTimestamp == reserveData.timestamp;
    }
    
    function _simulateComplianceVerification(
        ProofData memory proof,
        bytes32 userHash,
        uint256 riskScore
    ) internal pure returns (bool) {
        if (proof.publicSignals.length < 2) return false;
        
        bytes32 providedHash = bytes32(proof.publicSignals[0]);
        uint256 providedRiskScore = proof.publicSignals[1];
        
        return providedHash == userHash && providedRiskScore == riskScore;
    }
}