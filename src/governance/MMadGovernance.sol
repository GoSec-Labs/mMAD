// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "../interfaces/IGovernance.sol";
import "../utils/AccessControl.sol";
import "../utils/ReentrancyGuard.sol";
import "../libraries/Math.sol";
import "../interfaces/IERC20Extended.sol";

/**
 * @title MMadGovernance
 * @dev Decentralized governance for MMad stablecoin protocol
 */
abstract contract MMadGovernance is IGovernance, AccessControl, ReentrancyGuard, IERC20Extended {
    using Math for uint256;
    using Errors for *;
    using Events for *;
    
    // Governance state
    uint256 private _proposalCounter;
    mapping(uint256 => ProposalData) private _proposals;
    mapping(uint256 => mapping(address => bool)) private _hasVoted;
    mapping(uint256 => mapping(address => VoteType)) private _votes;
    
    // Governance parameters
    uint256 private _votingDelay = 1 days;      // 1 day before voting starts
    uint256 private _votingPeriod = 7 days;     // 7 days voting period
    uint256 private _proposalThreshold = 100000 * 10**18; // 100k tokens to propose
    uint256 private _quorumNumerator = 4;       // 4% quorum
    
    // Token and timelock integration
    IERC20Extended private _governanceToken;
    address private _timelock;
    
    // Events
    // event VotingDelaySet(uint256 oldVotingDelay, uint256 newVotingDelay);
    // event VotingPeriodSet(uint256 oldVotingPeriod, uint256 newVotingPeriod);
    // event ProposalThresholdSet(uint256 oldProposalThreshold, uint256 newProposalThreshold);
    event QuorumNumeratorUpdated(uint256 oldQuorumNumerator, uint256 newQuorumNumerator);
    event TimelockChange(address oldTimelock, address newTimelock);
    
    constructor(
        address governanceToken,
        address timelock,
        address admin
    ) {
        if (governanceToken == address(0)) revert Errors.ZeroAddress();
        if (timelock == address(0)) revert Errors.ZeroAddress();
        if (admin == address(0)) revert Errors.ZeroAddress();
        
        _governanceToken = IERC20Extended(governanceToken);
        _timelock = timelock;
        
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(DEFAULT_ADMIN_ROLE, timelock); // Timelock can execute admin functions
    }
    
    // Proposal functions
    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    ) public override returns (uint256) {
        if (getVotes(msg.sender, block.number - 1) < _proposalThreshold) {
            revert Errors.InsufficientVotingPower();
        }
        
        if (targets.length != values.length || targets.length != calldatas.length) {
            revert Errors.InvalidParameter();
        }
        
        if (targets.length == 0) revert Errors.InvalidParameter();
        
        uint256 proposalId = ++_proposalCounter;
        uint256 startBlock = block.number + _votingDelay;
        uint256 endBlock = startBlock + _votingPeriod;
        
        ProposalData storage proposal = _proposals[proposalId];
        proposal.id = proposalId;
        proposal.proposer = msg.sender;
        proposal.startBlock = startBlock;
        proposal.endBlock = endBlock;
        proposal.description = description;
        
        emit ProposalCreated(
            proposalId,
            msg.sender,
            targets,
            values,
            new string[](targets.length), // signatures array (empty for this implementation)
            calldatas,
            startBlock,
            endBlock,
            description
        );
        
        return proposalId;
    }
    
    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public payable override returns (uint256) {
        uint256 proposalId = hashProposal(targets, values, calldatas, descriptionHash);
        
        ProposalState currentState = state(proposalId);
        if (currentState != ProposalState.Succeeded && currentState != ProposalState.Queued) {
            revert Errors.ProposalNotActive();
        }
        
        ProposalData storage proposal = _proposals[proposalId];
        if (proposal.executed) revert Errors.ProposalAlreadyExecuted();
        
        proposal.executed = true;
        
        // Execute through timelock if set
        if (_timelock != address(0)) {
            _executeWithTimelock(targets, values, calldatas);
        } else {
            _executeDirect(targets, values, calldatas);
        }
        
        emit ProposalExecuted(proposalId);
        return proposalId;
    }
    
    function cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public override returns (uint256) {
        uint256 proposalId = hashProposal(targets, values, calldatas, descriptionHash);
        ProposalData storage proposal = _proposals[proposalId];
        
        if (proposal.executed) revert Errors.ProposalAlreadyExecuted();
        if (proposal.canceled) revert Errors.ProposalAlreadyCanceled();
        
        // Only proposer or admin can cancel
        if (msg.sender != proposal.proposer && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert Errors.Unauthorized();
        }
        
        proposal.canceled = true;
        emit ProposalCanceled(proposalId);
        
        return proposalId;
    }
    
    // Voting functions
    function castVote(uint256 proposalId, uint8 support) public override returns (uint256) {
        return _castVote(proposalId, msg.sender, VoteType(support), "");
    }
    
    function castVoteWithReason(
        uint256 proposalId,
        uint8 support,
        string calldata reason
    ) public override returns (uint256) {
        return _castVote(proposalId, msg.sender, VoteType(support), reason);
    }
    
    function castVoteBySig(
        uint256 proposalId,
        uint8 support,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public override returns (uint256) {
        // Implement signature verification for meta-transactions
        // Simplified for this example
        address voter = msg.sender; // In production, recover from signature
        return _castVote(proposalId, voter, VoteType(support), "");
    }
    
    // View functions
    function getProposal(uint256 proposalId) public view override returns (ProposalData memory) {
        return _proposals[proposalId];
    }
    
    function state(uint256 proposalId) public view override returns (ProposalState) {
        ProposalData storage proposal = _proposals[proposalId];
        
        if (proposal.executed) return ProposalState.Executed;
        if (proposal.canceled) return ProposalState.Canceled;
        
        uint256 currentBlock = block.number;
        
        if (currentBlock < proposal.startBlock) return ProposalState.Pending;
        if (currentBlock <= proposal.endBlock) return ProposalState.Active;
        
        uint256 totalVotes = proposal.forVotes + proposal.againstVotes + proposal.abstainVotes;
        
        if (totalVotes < quorum(proposal.startBlock)) return ProposalState.Defeated;
        if (proposal.forVotes <= proposal.againstVotes) return ProposalState.Defeated;
        
        return ProposalState.Succeeded;
    }
    
    function proposalThreshold() public view override returns (uint256) {
        return _proposalThreshold;
    }
    
    function quorum(uint256 blockNumber) public view override returns (uint256) {
        uint256 totalSupply = _governanceToken.totalSupply();
        return (totalSupply * _quorumNumerator) / 100;
    }
    
    function votingDelay() public view override returns (uint256) {
        return _votingDelay;
    }
    
    function votingPeriod() public view override returns (uint256) {
        return _votingPeriod;
    }
    
    function hasVoted(uint256 proposalId, address account) public view override returns (bool) {
        return _hasVoted[proposalId][account];
    }
    
    function getVotes(address account, uint256 blockNumber) public view override returns (uint256) {
        // In production, this would check historical balance/delegation at blockNumber
        return _governanceToken.balanceOf(account);
    }
    
    // Configuration functions
    function setVotingDelay(uint256 newVotingDelay) public override onlyRole(DEFAULT_ADMIN_ROLE) {
        uint256 oldVotingDelay = _votingDelay;
        _votingDelay = newVotingDelay;
        emit VotingDelaySet(oldVotingDelay, newVotingDelay);
    }
    
    function setVotingPeriod(uint256 newVotingPeriod) public override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newVotingPeriod == 0) revert Errors.InvalidParameter();
        uint256 oldVotingPeriod = _votingPeriod;
        _votingPeriod = newVotingPeriod;
        emit VotingPeriodSet(oldVotingPeriod, newVotingPeriod);
    }
    
    function setProposalThreshold(uint256 newProposalThreshold) public override onlyRole(DEFAULT_ADMIN_ROLE) {
        uint256 oldProposalThreshold = _proposalThreshold;
        _proposalThreshold = newProposalThreshold;
        emit ProposalThresholdSet(oldProposalThreshold, newProposalThreshold);
    }
    
    function updateTimelock(address newTimelock) public override onlyRole(DEFAULT_ADMIN_ROLE) {
        address oldTimelock = _timelock;
        _timelock = newTimelock;
        
        if (oldTimelock != address(0)) {
            _revokeRole(DEFAULT_ADMIN_ROLE, oldTimelock);
        }
        if (newTimelock != address(0)) {
            _grantRole(DEFAULT_ADMIN_ROLE, newTimelock);
        }
        
        emit TimelockChange(oldTimelock, newTimelock);
    }
    
    // Internal functions
    function _castVote(
        uint256 proposalId,
        address account,
        VoteType support,
        string memory reason
    ) internal returns (uint256) {
        ProposalData storage proposal = _proposals[proposalId];
        
        if (state(proposalId) != ProposalState.Active) revert Errors.ProposalNotActive();
        if (_hasVoted[proposalId][account]) revert Errors.AlreadyVoted();
        
        uint256 weight = getVotes(account, proposal.startBlock);
        if (weight == 0) revert Errors.InsufficientVotingPower();
        
        _hasVoted[proposalId][account] = true;
        _votes[proposalId][account] = support;
        
        if (support == VoteType.Against) {
            proposal.againstVotes += weight;
        } else if (support == VoteType.For) {
            proposal.forVotes += weight;
        } else {
            proposal.abstainVotes += weight;
        }
        
        emit VoteCast(account, proposalId, uint8(support), weight, reason);
        
        return weight;
    }
    
    function _executeWithTimelock(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas
    ) internal {
        // Execute through timelock contract
        for (uint256 i = 0; i < targets.length; i++) {
            (bool success, ) = _timelock.call(
                abi.encodeWithSignature(
                    "execute(address,uint256,bytes)",
                    targets[i],
                    values[i],
                    calldatas[i]
                )
            );
            require(success, "Timelock execution failed");
        }
    }
    
    function _executeDirect(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas
    ) internal {
        for (uint256 i = 0; i < targets.length; i++) {
            (bool success, ) = targets[i].call{value: values[i]}(calldatas[i]);
            require(success, "Direct execution failed");
        }
    }
    
    function hashProposal(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public pure returns (uint256) {
        return uint256(keccak256(abi.encode(targets, values, calldatas, descriptionHash)));
    }
}
