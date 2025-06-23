// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "./AccessControl.sol";

/**
 * @title Pausable
 * @dev Contract module which allows children to implement an emergency stop
 */
contract Pausable is AccessControl {
    bool private _paused;
    
    event Paused(address account);
    event Unpaused(address account);
    
    constructor() {
        _paused = false;
    }
    
    modifier whenNotPaused() {
        _requireNotPaused();
        _;
    }
    
    modifier whenPaused() {
        _requirePaused();
        _;
    }
    
    function paused() public view returns (bool) {
        return _paused;
    }
    
    function _requireNotPaused() internal view {
        if (paused()) revert Errors.Paused();
    }
    
    function _requirePaused() internal view {
        if (!paused()) revert Errors.NotPaused();
    }
    
    // FIX: Add 'virtual' keyword to both functions
    function pause() public virtual onlyRole(PAUSER_ROLE) {
        _pause();
    }
    
    function unpause() public virtual onlyRole(PAUSER_ROLE) {
        _unpause();
    }
    
    function _pause() internal whenNotPaused {
        _paused = true;
        emit Paused(msg.sender);
    }
    
    function _unpause() internal whenPaused {
        _paused = false;
        emit Unpaused(msg.sender);
    }
}