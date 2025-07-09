// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "../libraries/Errors.sol";


/**
 * @title ReentrancyGuard
 * @dev Contract module that helps prevent reentrant calls
 */
contract ReentrancyGuard {
    uint256 private constant _NOT_ENTERED = 1;
    uint256 private constant _ENTERED = 2;
    
    uint256 private _status;
    
    constructor() {
        _status = _NOT_ENTERED;
    }
    
    modifier nonReentrant() {
        _nonReentrantBefore();
        _;
        _nonReentrantAfter();
    }
    
    function _nonReentrantBefore() private {
        if (_status == _ENTERED) revert Errors.ReentrancyDetected();
        _status = _ENTERED;
    }
    
    function _nonReentrantAfter() private {
        _status = _NOT_ENTERED;
    }
    
    function _reentrancyGuardEntered() internal view returns (bool) {
        return _status == _ENTERED;
    }
}