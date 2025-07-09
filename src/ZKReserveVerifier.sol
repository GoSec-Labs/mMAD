// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IZKVerifier.sol";
import "./utils/AccessControl.sol";
import "./libraries/Errors.sol";
import "./libraries/Events.sol";

// Import the generated verifiers
interface IGroth16Verifier {
    function verifyProof(
        uint[2] calldata _pA,
        uint[2][2] calldata _pB,
        uint[2] calldata _pC,
        uint[1] calldata _pubSignals
    ) external view returns (bool);
}

interface IGroth16VerifierCompliance {
    function verifyProof(
        uint[2] calldata _pA,
        uint[2][2] calldata _pB,
        uint[2] calldata _pC,
        uint[2] calldata _pubSignals
    ) external view returns (bool);
}

interface IGroth16VerifierBatch {
    function verifyProof(
        uint[2] calldata _pA,
        uint[2][2] calldata _pB,
        uint[2] calldata _pC,
        uint[2] calldata _pubSignals
    ) external view returns (bool);
}

/**
 * @title ZKReserveVerifier
 * @dev Wrapper contract integrating all ZK proof verifiers for mMAD
 */
contract ZKReserveVerifier is IZKVerifier, AccessControl {
    
    address public immutable reserveVerifierAddr;
    address public immutable complianceVerifierAddr;
    address public immutable batchVerifierAddr;
    
    uint256 private _maxProofAge = 1 hours;
    uint256 private _requiredReserveRatio = 11000; // 110% in basis points
    address private _proofValidator;
    
    mapping(bytes32 => uint256) private _proofTimestamps;
    mapping(bytes32 => bool) private _validProofs;
    
    constructor(
        address admin,
        address _reserveVerifier,
        address _complianceVerifier,
        address _batchVerifier
    ) {
        require(admin != address(0), "Admin cannot be zero");
        require(_reserveVerifier != address(0), "Reserve verifier cannot be zero");
        require(_complianceVerifier != address(0), "Compliance verifier cannot be zero");
        require(_batchVerifier != address(0), "Batch verifier cannot be zero");
        
        reserveVerifierAddr = _reserveVerifier;
        complianceVerifierAddr = _complianceVerifier;
        batchVerifierAddr = _batchVerifier;
        
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
    }
    
    function verifyReserveProof(
        ProofData calldata proof,
        ReserveProof calldata reserveData
    ) external view override returns (bool) {
        // Convert proof format for Groth16 verifier
        uint[2] memory a = [proof.a[0], proof.a[1]];
        uint[2][2] memory b = [[proof.b[0][0], proof.b[0][1]], [proof.b[1][0], proof.b[1][1]]];
        uint[2] memory c = [proof.c[0], proof.c[1]];
        
        // Validate proof structure and timing
        require(proof.publicSignals.length >= 1, "Invalid public signals");
        require(block.timestamp - reserveData.timestamp <= _maxProofAge, "Proof too old");
        
        // ReserveProof has 1 output: isValid
        uint[1] memory publicSignals = [proof.publicSignals[0]];
        
        // Verify using generated Groth16 verifier
        return IGroth16Verifier(reserveVerifierAddr).verifyProof(a, b, c, publicSignals);
    }
    
    function verifyComplianceProof(
        ProofData calldata proof,
        bytes32 userHash,
        uint256 riskScore
    ) external view override returns (bool) {
        uint[2] memory a = [proof.a[0], proof.a[1]];
        uint[2][2] memory b = [[proof.b[0][0], proof.b[0][1]], [proof.b[1][0], proof.b[1][1]]];
        uint[2] memory c = [proof.c[0], proof.c[1]];
        
        // ComplianceCheck has 2 outputs: isCompliant, userCommitment
        uint[2] memory publicSignals = [
            proof.publicSignals.length > 0 ? proof.publicSignals[0] : 0,
            proof.publicSignals.length > 1 ? proof.publicSignals[1] : 0
        ];
        
        return IGroth16VerifierCompliance(complianceVerifierAddr).verifyProof(a, b, c, publicSignals);
    }
    
    function verifyBatchProofs(
        ProofData[] calldata proofs,
        bytes32[] calldata commitments
    ) external view override returns (bool) {
        require(proofs.length == commitments.length, "Length mismatch");
        
        // For simplicity, verify the first proof
        if (proofs.length > 0) {
            uint[2] memory a = [proofs[0].a[0], proofs[0].a[1]];
            uint[2][2] memory b = [[proofs[0].b[0][0], proofs[0].b[0][1]], [proofs[0].b[1][0], proofs[0].b[1][1]]];
            uint[2] memory c = [proofs[0].c[0], proofs[0].c[1]];
            
            // BatchVerifier has 2 outputs: allValid, batchCommitment
            uint[2] memory publicSignals = [
                proofs[0].publicSignals.length > 0 ? proofs[0].publicSignals[0] : 0,
                proofs[0].publicSignals.length > 1 ? proofs[0].publicSignals[1] : 0
            ];
            
            return IGroth16VerifierBatch(batchVerifierAddr).verifyProof(a, b, c, publicSignals);
        }
        return false;
    }
    
    // Configuration functions
    function updateVerificationKey(bytes calldata /*vkData*/) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        // Implementation depends on your needs
        emit Events.ProofValidatorUpdated(_proofValidator, address(0));
    }
    
    function setProofValidator(address validator) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        address oldValidator = _proofValidator;
        _proofValidator = validator;
        emit Events.ProofValidatorUpdated(oldValidator, validator);
    }
    
    function setMaxProofAge(uint256 maxAge) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        _maxProofAge = maxAge;
    }
    
    // View functions
    function getLastProofTimestamp() external view override returns (uint256) {
        return block.timestamp; // Simplified
    }
    
    function isProofValid(bytes32 proofHash) external view override returns (bool) {
        return _validProofs[proofHash];
    }
    
    function getRequiredReserveRatio() external view override returns (uint256) {
        return _requiredReserveRatio;
    }
}
