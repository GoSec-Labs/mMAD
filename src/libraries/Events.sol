// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title Events
 * @dev Centralized event definitions for consistent logging
 */
library Events{
    // Token Events
    event TokenMinted(address indexed to, uint256 amount, bytes32 proofHash);
    event TokenBurned(address indexed from, uint256 amount);
    event ReservesUpdated(uint256 newAmount, uint256 backingRatio);
    event BackingRatioUpdated(uint256 oldRatio, uint256 newRatio);
    
    // ZK Proof Events
    event ProofVerified(bytes32 indexed proofHash, address indexed verifier);
    event ReserveProofSubmitted(uint256 reserves, uint256 supply, uint256 ratio);
    event ComplianceProofVerified(bytes32 indexed userHash, uint256 riskScore);
    event ProofValidatorUpdated(address oldValidator, address newValidator);
    
    // Access Control Events
    event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender);
    event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender);
    event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole);
    
    // Governance Events
    event ProposalCreated(uint256 indexed proposalId, address indexed proposer, string description);
    event VoteCast(address indexed voter, uint256 indexed proposalId, uint8 support, uint256 weight);
    event ProposalExecuted(uint256 indexed proposalId, bytes32 indexed descriptionHash);
    event ProposalQueued(uint256 indexed proposalId, uint256 executeTime);
    
    // Security Events
    event EmergencyPause(address indexed account);
    event EmergencyUnpause(address indexed account);
    event EmergencyWithdrawal(address indexed token, uint256 amount, address indexed to);

    // ERC20 Events
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    // Mint and Burn Events
    event Mint(address indexed to, uint256 amount);
    event Burn(address indexed from, uint256 amount);
}