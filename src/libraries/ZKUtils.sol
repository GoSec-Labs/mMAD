// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title ZKUtils
 * @dev Utilities for zero-knowledge proof handling
 */
library ZKUtils {
    struct ProofData {
        uint256[2] a;
        uint256[2][2] b;
        uint256[2] c;
        uint256[] publicSignals;
    }
    
    struct ReserveProof {
        uint256 minRequiredReserve;
        uint256 currentSupply;
        uint256 timestamp;
        bytes32 commitment;
    }
    
    uint256 internal constant MAX_PROOF_AGE = 1 hours;
    
    /**
     * @dev Generate proof hash for tracking
     */
    function generateProofHash(ProofData memory proof) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(
            proof.a[0], proof.a[1],
            proof.b[0][0], proof.b[0][1], proof.b[1][0], proof.b[1][1],
            proof.c[0], proof.c[1],
            proof.publicSignals
        ));
    }
    
    /**
     * @dev Validate proof structure
     */
    function validateProofStructure(ProofData memory proof) internal pure returns (bool) {
        return proof.a.length == 2 &&
               proof.b.length == 2 &&
               proof.b[0].length == 2 &&
               proof.b[1].length == 2 &&
               proof.c.length == 2 &&
               proof.publicSignals.length > 0;
    }
    
    /**
     * @dev Check if proof is within valid time window
     */
    function isProofTimestampValid(uint256 proofTimestamp) internal view returns (bool) {
        return block.timestamp - proofTimestamp <= MAX_PROOF_AGE;
    }
    
    /**
     * @dev Validate reserve proof data
     */
    function validateReserveProof(ReserveProof memory reserveData) internal view returns (bool) {
        return reserveData.minRequiredReserve > 0 &&
               reserveData.currentSupply > 0 &&
               isProofTimestampValid(reserveData.timestamp);
    }
    
    /**
     * @dev Extract public signals for reserve proof
     */
    function extractReserveSignals(uint256[] memory publicSignals) internal pure returns (
        uint256 minRequired,
        uint256 currentSupply,
        uint256 timestamp
    ) {
        require(publicSignals.length >= 3, "ZKUtils: insufficient public signals");
        return (publicSignals[0], publicSignals[1], publicSignals[2]);
    }
}