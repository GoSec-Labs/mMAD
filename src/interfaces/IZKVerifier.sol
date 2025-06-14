// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

interface IZKVerifier {
    struct ProofData {
        uint256[2] a; //point a
        uint256[2][2] b; // point b 
        uint256[2] c; // point c
        uint256[] publicSignals; //public inputs 
    }

    struct ReserveProof {
        uint256 minRequiredReserved;
        uint256 currentSupply;
        uint256 timestamp;
        bytes32 commitemnet;
    }

    //core function 
    function verifyReserveProof(
        ProofData calldata reserveData
    )external view returns (bool);

    function verifyComplianceProof(
        ProofData calldata proof,
        bytes32 userHash,
        uint256 riskScore
    ) external view returns (bool);
    
    function verifyBatchProofs(
        ProofData[] calldata proofs,
        bytes32[] calldata commitments
    ) external view returns (bool);
    
    // Configuration functions
    function updateVerificationKey(bytes calldata vkData) external;
    function setProofValidator(address validator) external;
    function setMaxProofAge(uint256 maxAge) external;
    
    // View functions
    function getLastProofTimestamp() external view returns (uint256);
    function isProofValid(bytes32 proofHash) external view returns (bool);
    function getRequiredReserveRatio() external view returns (uint256);
    
    // Events
    event ProofVerified(
        bytes32 indexed proofHash,
        address indexed verifier,
        uint256 timestamp
    );
    event ReserveProofSubmitted(
        uint256 minRequired,
        uint256 currentSupply,
        uint256 timestamp
    );
    event ComplianceProofVerified(
        bytes32 indexed userHash,
        uint256 riskScore,
        uint256 timestamp
    );
    event VerificationKeyUpdated(bytes32 keyHash);
}