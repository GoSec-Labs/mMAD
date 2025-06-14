// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "./interfaces/IZKVerifier.sol";
import "./utils/AccessControl.sol";

/**
 * @title ZKReserveVerifier
 * @dev Verifies zero-knowledge proofs for reserve backing and compliance
 */
contract ZKReserveVerifier is IZKVerifier, AccessControl, Pausable {
    using ZKUtils for *;
    using Errors for *;
    using Events for *;
    
    // Verification keys (simplified - in production would use actual Groth16 keys)
    bytes32 private _reserveVerificationKey;
    bytes32 private _complianceVerificationKey;
    
    // Proof tracking
    mapping(bytes32 => bool) private _verifiedProofs;
    mapping(bytes32 => uint256) private _proofTimestamps;
    
    // Configuration
    uint256 private _maxProofAge = 1 hours;
    uint256 private _requiredReserveRatio = 110 * Math.PERCENTAGE_SCALE / 100; // 110%
    address private _proofValidator;
    
    event VerificationKeyUpdated(bytes32 indexed keyType, bytes32 keyHash);
    event ProofValidatorUpdated(address oldValidator, address newValidator);
    event MaxProofAgeUpdated(uint256 oldAge, uint256 newAge);
    
    constructor(address admin) {
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(PROOF_VALIDATOR_ROLE, admin);
        _proofValidator = admin;
    }
    
    function verifyReserveProof(
        ProofData calldata proof,
        ReserveProof calldata reserveData
    ) external view override returns (bool) {
        // Validate proof structure
        if (!ZKUtils.validateProofStructure(proof)) return false;
        if (!ZKUtils.validateReserveProof(reserveData)) return false;
        
        // Verify reserve ratio meets minimum requirements
        uint256 actualRatio = Math.calculateBackingRatio(
            reserveData.minRequiredReserve,
            reserveData.currentSupply
        );
        
        if (actualRatio < _requiredReserveRatio) return false;
        
        // Verify proof timestamp is recent
        if (!ZKUtils.isProofTimestampValid(reserveData.timestamp)) return false;
        
        // In production, this would verify the actual Groth16 proof
        // For now, we simulate verification based on structure and data validation
        return _simulateGroth16Verification(proof, reserveData);
    }
    
    function verifyComplianceProof(
        ProofData calldata proof,
        bytes32 userHash,
        uint256 riskScore
    ) external view override returns (bool) {
        if (!ZKUtils.validateProofStructure(proof)) return false;
        if (userHash == bytes32(0)) return false;
        if (riskScore > 100) return false; // Risk score 0-100
        
        // In production, verify actual compliance ZK proof
        return _simulateComplianceVerification(proof, userHash, riskScore);
    }
    
    function verifyBatchProofs(
        ProofData[] calldata proofs,
        bytes32[] calldata commitments
    ) external view override returns (bool) {
        if (proofs.length != commitments.length) return false;
        if (proofs.length == 0) return false;
        
        for (uint256 i = 0; i < proofs.length; i++) {
            if (!ZKUtils.validateProofStructure(proofs[i])) return false;
            if (commitments[i] == bytes32(0)) return false;
        }
        
        // Batch verification would be more efficient in production
        return true;
    }
    
    // Configuration functions
    function updateVerificationKey(bytes calldata vkData) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        bytes32 keyHash = keccak256(vkData);
        _reserveVerificationKey = keyHash;
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
        return block.timestamp; // In production, track last verified proof
    }
    
    function isProofValid(bytes32 proofHash) external view override returns (bool) {
        return _verifiedProofs[proofHash] && 
               block.timestamp - _proofTimestamps[proofHash] <= _maxProofAge;
    }
    
    function getRequiredReserveRatio() external view override returns (uint256) {
        return _requiredReserveRatio;
    }
    
    // Internal verification functions (simplified for demo)
    function _simulateGroth16Verification(
        ProofData memory proof,
        ReserveProof memory reserveData
    ) internal pure returns (bool) {
        // In production, this would use actual pairing checks
        // For now, verify that public signals match expected format
        if (proof.publicSignals.length < 3) return false;
        
        uint256 providedMinReserve = proof.publicSignals[0];
        uint256 providedSupply = proof.publicSignals[1];
        uint256 providedTimestamp = proof.publicSignals[2];
        
        return providedMinReserve == reserveData.minRequiredReserve &&
               providedSupply == reserveData.currentSupply &&
               providedTimestamp == reserveData.timestamp;
    }
    
    function _simulateComplianceVerification(
        ProofData memory proof,
        bytes32 userHash,
        uint256 riskScore
    ) internal pure returns (bool) {
        // Simplified compliance verification
        if (proof.publicSignals.length < 2) return false;
        
        // Verify user hash is included in public signals
        bytes32 providedHash = bytes32(proof.publicSignals[0]);
        uint256 providedRiskScore = proof.publicSignals[1];
        
        return providedHash == userHash && providedRiskScore == riskScore;
    }
}
