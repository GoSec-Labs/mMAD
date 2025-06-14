// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "./utils/AccessControl.sol";

/**
 * @title Timelock
 * @dev Timelock controller for delayed execution of governance proposals
 */
contract Timelock is AccessControl {
    using Errors for *;
    using Events for *;
    
    bytes32 public constant PROPOSER_ROLE = keccak256("PROPOSER_ROLE");
    bytes32 public constant EXECUTOR_ROLE = keccak256("EXECUTOR_ROLE");
    bytes32 public constant CANCELLER_ROLE = keccak256("CANCELLER_ROLE");
    
    uint256 private _minDelay;
    mapping(bytes32 => bool) private _timestamps;
    
    event CallScheduled(
        bytes32 indexed id,
        uint256 indexed index,
        address target,
        uint256 value,
        bytes data,
        bytes32 predecessor,
        uint256 delay
    );
    
    event CallExecuted(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data);
    event CallCancelled(bytes32 indexed id);
    event MinDelayChange(uint256 oldDuration, uint256 newDuration);
    
    constructor(
        uint256 minDelay,
        address[] memory proposers,
        address[] memory executors,
        address admin
    ) {
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setRoleAdmin(PROPOSER_ROLE, DEFAULT_ADMIN_ROLE);
        _setRoleAdmin(EXECUTOR_ROLE, DEFAULT_ADMIN_ROLE);
        _setRoleAdmin(CANCELLER_ROLE, DEFAULT_ADMIN_ROLE);
        
        // Grant roles to initial accounts
        for (uint256 i = 0; i < proposers.length; ++i) {
            _setupRole(PROPOSER_ROLE, proposers[i]);
            _setupRole(CANCELLER_ROLE, proposers[i]);
        }
        for (uint256 i = 0; i < executors.length; ++i) {
            _setupRole(EXECUTOR_ROLE, executors[i]);
        }
        
        _minDelay = minDelay;
        emit MinDelayChange(0, minDelay);
    }
    
    modifier onlyRoleOrOpenRole(bytes32 role) {
        if (!hasRole(role, address(0))) {
            _checkRole(role);
        }
        _;
    }
    
    receive() external payable {}
    
    function isOperation(bytes32 id) public view returns (bool) {
        return getTimestamp(id) > 0;
    }
    
    function isOperationPending(bytes32 id) public view returns (bool) {
        return getTimestamp(id) > block.timestamp;
    }
    
    function isOperationReady(bytes32 id) public view returns (bool) {
        uint256 timestamp = getTimestamp(id);
        return timestamp > 0 && timestamp <= block.timestamp;
    }
    
    function isOperationDone(bytes32 id) public view returns (bool) {
        return getTimestamp(id) == 1;
    }
    
    function getTimestamp(bytes32 id) public view returns (uint256 timestamp) {
        return _timestamps[id];
    }
    
    function getMinDelay() public view returns (uint256 duration) {
        return _minDelay;
    }
    
    function hashOperation(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt
    ) public pure returns (bytes32 hash) {
        return keccak256(abi.encode(target, value, data, predecessor, salt));
    }
    
    function hashOperationBatch(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata payloads,
        bytes32 predecessor,
        bytes32 salt
    ) public pure returns (bytes32 hash) {
        return keccak256(abi.encode(targets, values, payloads, predecessor, salt));
    }
    
    function schedule(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt,
        uint256 delay
    ) public onlyRole(PROPOSER_ROLE) {
        bytes32 id = hashOperation(target, value, data, predecessor, salt);
        _schedule(id, delay);
        emit CallScheduled(id, 0, target, value, data, predecessor, delay);
    }
    
    function scheduleBatch(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata payloads,
        bytes32 predecessor,
        bytes32 salt,
        uint256 delay
    ) public onlyRole(PROPOSER_ROLE) {
        if (targets.length != values.length || targets.length != payloads.length) {
            revert Errors.InvalidParameter();
        }
        
        bytes32 id = hashOperationBatch(targets, values, payloads, predecessor, salt);
        _schedule(id, delay);
        
        for (uint256 i = 0; i < targets.length; ++i) {
            emit CallScheduled(id, i, targets[i], values[i], payloads[i], predecessor, delay);
        }
    }
    
    function cancel(bytes32 id) public onlyRole(CANCELLER_ROLE) {
        if (!isOperationPending(id)) revert Errors.TimelockNotReady();
        delete _timestamps[id];
        
        emit CallCancelled(id);
    }
    
    function execute(
        address target,
        uint256 value,
        bytes calldata payload,
        bytes32 predecessor,
        bytes32 salt
    ) public payable onlyRoleOrOpenRole(EXECUTOR_ROLE) {
        bytes32 id = hashOperation(target, value, payload, predecessor, salt);
        
        _beforeCall(id, predecessor);
        _execute(target, value, payload);
        emit CallExecuted(id, 0, target, value, payload);
        _afterCall(id);
    }
    
    function executeBatch(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata payloads,
        bytes32 predecessor,
        bytes32 salt
    ) public payable onlyRoleOrOpenRole(EXECUTOR_ROLE) {
        if (targets.length != values.length || targets.length != payloads.length) {
            revert Errors.InvalidParameter();
        }
        
        bytes32 id = hashOperationBatch(targets, values, payloads, predecessor, salt);
        
        _beforeCall(id, predecessor);
        for (uint256 i = 0; i < targets.length; ++i) {
            address target = targets[i];
            uint256 value = values[i];
            bytes calldata payload = payloads[i];
            _execute(target, value, payload);
            emit CallExecuted(id, i, target, value, payload);
        }
        _afterCall(id);
    }
    
    function updateDelay(uint256 newDelay) external {
        if (msg.sender != address(this)) revert Errors.Unauthorized();
        emit MinDelayChange(_minDelay, newDelay);
        _minDelay = newDelay;
    }
    
    function _schedule(bytes32 id, uint256 delay) private {
        if (isOperation(id)) revert Errors.InvalidParameter();
        if (delay < getMinDelay()) revert Errors.InvalidParameter();
        
        _timestamps[id] = block.timestamp + delay;
    }
    
    function _beforeCall(bytes32 id, bytes32 predecessor) private view {
        if (!isOperationReady(id)) revert Errors.TimelockNotReady();
        if (predecessor != bytes32(0) && !isOperationDone(predecessor)) {
            revert Errors.InvalidParameter();
        }
    }
    
    function _afterCall(bytes32 id) private {
        if (!isOperationReady(id)) revert Errors.TimelockNotReady();
        _timestamps[id] = 1;
    }
    
    function _execute(address target, uint256 value, bytes calldata data) private {
        (bool success, ) = target.call{value: value}(data);
        require(success, "Timelock: underlying transaction reverted");
    }
}