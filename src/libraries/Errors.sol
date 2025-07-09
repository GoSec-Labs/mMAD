// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title Errors
 * @dev Custom error definitions for gas-efficient reverts
 */
library Errors {
    // Access Control Errors
    error Unauthorized();
    error InvalidRole();
    error RoleAlreadyGranted();
    error RoleNotFound();
    
    // Token Errors
    error InsufficientBalance();
    error InsufficientAllowance();
    error InvalidAmount();
    error ExceedsMaxSupply();
    error MintingPaused();
    error BurningPaused();
    error TransferToZeroAddress();
    error TransferFromZeroAddress();
    error ApproveToZeroAddress();
    error ApproveFromZeroAddress();
    
    // ZK Proof Errors
    error InvalidProof();
    error ProofExpired();
    error ProofAlreadyUsed();
    error InvalidReserveProof();
    error InvalidComplianceProof();
    error InsufficientReserves();
    error ReserveRatioTooLow();
    error InvalidVerificationKey();
    
    // Governance Errors
    error ProposalNotFound();
    error ProposalNotActive();
    error ProposalAlreadyExecuted();
    error ProposalAlreadyCanceled();
    error VotingPeriodEnded();
    error VotingPeriodNotEnded();
    error QuorumNotReached();
    error InvalidVoteType();
    error AlreadyVoted();
    error InsufficientVotingPower();
    
    // General Errors
    error Paused();
    error NotPaused();
    error ZeroAddress();
    error InvalidParameter();
    error TimelockNotReady();
    error ReentrancyDetected();
    error InvalidSignature();
}